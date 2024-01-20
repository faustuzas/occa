package di

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"

	"go.uber.org/zap"
)

var (
	ErrDependencyNotFound = fmt.Errorf("not found")
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

type Starter interface {
	Start() error
}

type Stopper interface {
	Stop() error
}

type Params struct {
	Logger *zap.Logger

	ConfigureInterruptTermination bool
	TerminationCh                 <-chan struct{}
}

func (p Params) Validate() error {
	if p.Logger == nil {
		return fmt.Errorf("logger must be provided")
	}
	return nil
}

type Application struct {
	p Params

	providers      map[reflect.Type]providerWrapper
	singletonCache map[reflect.Type]reflect.Value
	stoppers       []Stopper
}

func NewApplication(params Params) (*Application, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validating parameters: %w", err)
	}

	return &Application{
		p: params,

		providers:      map[reflect.Type]providerWrapper{},
		singletonCache: map[reflect.Type]reflect.Value{},
	}, nil
}

func (a *Application) MultiRegister(providers ...any) error {
	for _, p := range providers {
		if err := a.Register(p); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) Register(provider any) error {
	providedType, err := resolveProvidedType(provider)
	if err != nil {
		return err
	}

	if _, ok := a.providers[providedType]; ok {
		return fmt.Errorf("provider for type %v already registered", providedType)
	}

	a.providers[providedType] = wrapProvider(provider)
	return nil
}

func resolveProvidedType(provider any) (reflect.Type, error) {
	providerType := reflect.TypeOf(provider)
	if providerType.Kind() != reflect.Func {
		return providerType, nil
	}

	for i := 0; i < providerType.NumIn(); i++ {
		in := providerType.In(i)
		if in == errorType {
			return nil, fmt.Errorf("provider function cannot have error as a dependency")
		}
	}

	if numOut := providerType.NumOut(); numOut > 2 || numOut == 0 {
		return nil, fmt.Errorf("provider function must return a value and an optional error")
	}

	firstArgType := providerType.Out(0)
	if firstArgType == errorType {
		return nil, fmt.Errorf("provider function first parameter cannot be error")
	}

	if providerType.NumOut() == 2 && providerType.Out(1) != errorType {
		return nil, fmt.Errorf("provider function second return value can be only error")
	}

	return firstArgType, nil
}

func (a *Application) Exec(f any) error {
	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return fmt.Errorf("function must be provided")
	}

	if fType.NumOut() != 0 {
		return fmt.Errorf("exec cannot have return values")
	}

	arguments := make([]reflect.Value, 0, fType.NumIn())
	for i := 0; i < fType.NumIn(); i++ {
		requiredType := fType.In(i)
		if requiredType == errorType {
			return fmt.Errorf("exec cannot have error as a dependency")
		}

		obj, err := a.provideForType(requiredType, depTracker{})
		if err != nil {
			return fmt.Errorf("providing type %v: %w", requiredType, err)
		}

		arguments = append(arguments, obj)
	}

	returned := reflect.ValueOf(f).Call(arguments)
	if len(returned) == 1 {
		return returned[0].Interface().(error)
	}
	return nil
}

func (a *Application) provideForType(requestedType reflect.Type, dt depTracker) (reflect.Value, error) {
	if val, ok := a.singletonCache[requestedType]; ok {
		return val, nil
	}

	if dt[requestedType] {
		return reflect.Value{}, fmt.Errorf("circular dependency detected")
	}
	dt[requestedType] = true

	val, err := a.provideNew(requestedType, dt)
	if err != nil {
		return reflect.Value{}, err
	}
	a.singletonCache[requestedType] = val

	if val.CanInterface() {
		if s, ok := val.Interface().(Starter); ok {
			if err = s.Start(); err != nil {
				return reflect.Value{}, fmt.Errorf("starting %v: %w", requestedType, err)
			}
		}

		if s, ok := val.Interface().(Stopper); ok {
			a.stoppers = append(a.stoppers, s)
		}
	}

	return val, nil
}

func (a *Application) provideNew(requestedType reflect.Type, dt depTracker) (reflect.Value, error) {
	for typ, provider := range a.providers {
		if typ == requestedType || typ.AssignableTo(requestedType) || typ.ConvertibleTo(requestedType) {
			value, err := a.provide(provider, dt)
			if err != nil {
				return reflect.Value{}, err
			}

			if requestedType.Kind() == reflect.Interface {
				interfaceValue := reflect.New(requestedType).Elem()
				interfaceValue.Set(value)

				value = interfaceValue
			}

			return value, nil
		}
	}

	return reflect.Value{}, ErrDependencyNotFound
}

func (a *Application) provide(pw providerWrapper, dt depTracker) (reflect.Value, error) {
	providerType := reflect.TypeOf(pw.provider)
	switch providerType.Kind() {
	case reflect.Struct, reflect.Pointer:
		return reflect.ValueOf(pw.provider), nil
	case reflect.Func:
		arguments := make([]reflect.Value, 0, providerType.NumIn())
		for i := 0; i < providerType.NumIn(); i++ {
			requiredType := providerType.In(i)
			obj, err := a.provideForType(requiredType, dt)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("providing type %v: %w", requiredType, err)
			}

			arguments = append(arguments, obj)
		}

		returned := reflect.ValueOf(pw.provider).Call(arguments)
		for len(returned) == 2 {
			err := returned[1]
			return reflect.Value{}, err.Elem().Interface().(error)
		}

		return returned[0], nil
	}
	return reflect.Value{}, fmt.Errorf("provider type %v not supported", providerType.Kind())
}

func (a *Application) WaitForTerminationAndClose() {
	var iCh <-chan struct{}
	if a.p.ConfigureInterruptTermination {
		iCh = interruptCh()
	}

	select {
	case <-iCh:
		a.p.Logger.Info("received interrupt, terminating...")
	case <-a.p.TerminationCh:
		a.p.Logger.Info("received a request to terminate from channel...")
	}

	for _, s := range a.stoppers {
		if err := s.Stop(); err != nil {
			a.p.Logger.Error("error while stopping service",
				zap.String("service", fmt.Sprintf("%T", s)),
				zap.Error(err))
		}
	}
}

func interruptCh() <-chan struct{} {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	resultCh := make(chan struct{})
	go func() {
		<-ch
		close(resultCh)
	}()

	return resultCh
}

func wrapProvider(p any) providerWrapper {
	return providerWrapper{
		provider: p,
	}
}

type providerWrapper struct {
	provider any
}

//func newDepTrackerFor(typ reflect.Type) *depTracker {
//
//}
//
//type depTracker struct {
//	trackers map[]
//}

// depTracker is used to track requested dependencies in the graph to prevent circles.
type depTracker map[reflect.Type]bool
