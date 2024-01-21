package di

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
)

var (
	ErrDependencyNotFound = fmt.Errorf("not found")
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

type StartFn func(errCh chan<- error, appClosedCh <-chan struct{}) error
type StopFn func(ctx context.Context) error

type stopper struct {
	id string
	fn StopFn
}

type Lifecycle struct {
	start StartFn
	stop  StopFn
}

func (l *Lifecycle) RegisterStarter(s StartFn) {
	l.start = s
}

func (l *Lifecycle) RegisterStopper(s StopFn) {
	l.stop = s
}

type Params struct {
	Logger *zap.Logger

	TerminationTimeout time.Duration

	CloseCh <-chan struct{}
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

	errCh    chan error
	closedCh chan struct{}

	stoppers []stopper
}

func NewApplication(params Params) (*Application, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validating parameters: %w", err)
	}

	return &Application{
		p: params,

		errCh:    make(chan error),
		closedCh: make(chan struct{}),

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

	if fType.NumOut() > 1 || (fType.NumOut() == 1 && fType.Out(0) != errorType) {
		return fmt.Errorf("function can have only optional return error")
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
		lf := &Lifecycle{}

		arguments := make([]reflect.Value, 0, providerType.NumIn())
		for i := 0; i < providerType.NumIn(); i++ {
			requiredType := providerType.In(i)

			if requiredType == reflect.TypeOf(lf) {
				arguments = append(arguments, reflect.ValueOf(lf))
				continue
			}

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

		result := returned[0]
		if lf.start != nil {
			if err := lf.start(a.errCh, a.closedCh); err != nil {
				return reflect.Value{}, fmt.Errorf("starting %v: %w", result.Type(), err)
			}
		}

		if lf.stop != nil {
			a.stoppers = append(a.stoppers, stopper{
				id: fmt.Sprintf("%v", result.Type()),
				fn: lf.stop,
			})
		}

		return result, nil
	}
	return reflect.Value{}, fmt.Errorf("provider type %v not supported", providerType.Kind())
}

func (a *Application) WaitForTermination() {
	select {
	case err := <-a.errCh:
		a.p.Logger.Error("received a critical error from a running service, terminating...", zap.Error(err))
	case <-a.p.CloseCh:
		a.p.Logger.Info("received a request to terminate...")
	}
}

func (a *Application) Close() {
	ctx := context.Background()
	if a.p.TerminationTimeout != 0 {
		c, cancelFn := context.WithTimeout(ctx, a.p.TerminationTimeout)
		defer cancelFn()

		ctx = c
	}

	for _, s := range a.stoppers {
		a.p.Logger.Debug("stopping service", zap.String("service", fmt.Sprintf("%T", s.id)))

		if err := s.fn(ctx); err != nil {
			a.p.Logger.Error("error while stopping service",
				zap.String("service", fmt.Sprintf("%v", s.id)),
				zap.Error(err))
		}
	}

	close(a.closedCh)
}

func (a *Application) WaitForTerminationAndClose() {
	a.WaitForTermination()
	a.Close()
}

func wrapProvider(p any) providerWrapper {
	return providerWrapper{
		provider: p,
	}
}

type providerWrapper struct {
	provider any
}

// depTracker is used to track requested dependencies in the graph to prevent circles.
type depTracker map[reflect.Type]bool
