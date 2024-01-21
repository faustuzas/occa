package di

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"

	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func TestParamsValidation(t *testing.T) {
	tests := map[string]struct {
		params    Params
		expectErr string
	}{
		"logger missing": {
			params:    Params{},
			expectErr: "validating parameters: logger must be provided",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app, err := NewApplication(tt.params)
			require.Nil(t, app)
			require.Equal(t, tt.expectErr, err.Error())
		})
	}
}

func TestDI_ErrorHandling(t *testing.T) {
	tests := map[string]struct {
		provider          any
		expectRegisterErr string

		execFn        any
		expectExecErr string
	}{
		"provider function can return one value": {
			provider: func() *aService {
				return nil
			},
			expectRegisterErr: "",
		},
		"provider function cannot return only error": {
			provider: func() error {
				return nil
			},
			expectRegisterErr: "provider function first parameter cannot be error",
		},
		"provider function can return error as a second value": {
			provider: func() (*aService, error) {
				return nil, nil
			},
			expectRegisterErr: "",
		},
		"provider function cannot have more than 2 return values": {
			provider: func() (*aService, error, error) {
				return nil, nil, nil
			},
			expectRegisterErr: "provider function must return a value and an optional error",
		},
		"provider function cannot return 2 non-error typed values": {
			provider: func() (*aService, *aService) {
				return nil, nil
			},
			expectRegisterErr: "provider function second return value can be only error",
		},
		"provider function cannot have error as dependency": {
			provider: func(error) *aService {
				return nil
			},
			expectRegisterErr: "provider function cannot have error as a dependency",
		},
		"provider requests provided type as dependency": {
			provider: func(*aService) *aService {
				return nil
			},
			execFn:        func(*aService) {},
			expectExecErr: "providing type *di.aService: providing type *di.aService: circular dependency detected",
		},
		"propagates error from register to exec": {
			provider: func() (*aService, error) {
				return nil, fmt.Errorf("failed to create")
			},
			execFn:        func(nc *aService) {},
			expectExecErr: "providing type *di.aService: failed to create",
		},
		"exec accepts only function": {
			execFn:        5,
			expectExecErr: "function must be provided",
		},
		"exec requests not registered dependency": {
			execFn:        func(p heap.Interface) {},
			expectExecErr: "providing type heap.Interface: not found",
		},
		"exec cannot have error as a dependency": {
			execFn:        func(error) {},
			expectExecErr: "exec cannot have error as a dependency",
		},
		"exec cannot have return values": {
			execFn:        func() *aService { return nil },
			expectExecErr: "function can have only optional return error",
		},
		"error from exec is propagated": {
			execFn:        func() error { return fmt.Errorf("broken") },
			expectExecErr: "broken",
		},
		"starter error is propagated to exec": {
			provider: func(lc *Lifecycle) cService {
				lc.RegisterStarter(func(_ chan<- error, _ <-chan struct{}) error {
					return fmt.Errorf("disk is broken")
				})
				return cService{}
			},
			execFn:        func(cService) {},
			expectExecErr: "providing type di.cService: starting di.cService: disk is broken",
		},
		"internal implementation not found if only interface was provided": {
			provider: func() SmallNum {
				return &aService{}
			},
			execFn:        func(service *aService) {},
			expectExecErr: "providing type *di.aService: not found",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			app := newTestApplication()

			if tt.provider != nil {
				rErr := app.Register(tt.provider)
				if tt.expectRegisterErr != "" {
					require.Equal(t, tt.expectRegisterErr, rErr.Error())
					return
				}
				require.NoError(t, rErr)
			}

			if tt.execFn != nil {
				eErr := app.Exec(tt.execFn)
				if tt.expectExecErr != "" {
					require.Error(t, eErr)
					require.Equal(t, tt.expectExecErr, eErr.Error())
					return
				}
				require.NoError(t, eErr)
			}
		})
	}
}

func TestDI_ResolveImplementation(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return &aService{n: 100}
	}))

	var resolvedNum int
	require.NoError(t, app.Exec(func(nc *aService) {
		resolvedNum = nc.n
	}))
	require.Equal(t, 100, resolvedNum)
}

func TestDI_ResolveInterfaceToRegisteredImplementation(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return &aService{n: 10}
	}))

	var resolvedValue string
	require.NoError(t, app.Exec(func(s fmt.Stringer) {
		resolvedValue = s.String()
	}))
	require.Equal(t, "10", resolvedValue)
}

func TestDI_ResolvesSingletonMultipleTimes(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return &aService{n: rand.Int() + 1}
	}))

	var firstRetrieval int
	require.NoError(t, app.Exec(func(a *aService) {
		firstRetrieval = a.n
	}))

	var secondRetrieval int
	require.NoError(t, app.Exec(func(a *aService) {
		secondRetrieval = a.n
	}))

	require.Equal(t, firstRetrieval, secondRetrieval)
}

func TestDI_ProviderWithDependencies(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return &aService{n: 100}
	}))

	require.NoError(t, app.Register(func(a *aService) bService {
		return bService{a: a}
	}))

	var resolvedNum int
	require.NoError(t, app.Exec(func(b bService) {
		resolvedNum = b.a.n
	}))
	require.Equal(t, 100, resolvedNum)
}

func TestDI_ProvideAndResolveStruct(t *testing.T) {
	app := newTestApplication()

	obj := aService{n: 100}

	require.NoError(t, app.Register(obj))

	var resolvedObj aService
	require.NoError(t, app.Exec(func(a aService) {
		resolvedObj = a
	}))
	require.NotEmpty(t, resolvedObj)
	require.Equal(t, obj, resolvedObj)
}

func TestDI_ProvideAndResolvePointer(t *testing.T) {
	app := newTestApplication()

	ptr := &aService{n: 100}

	require.NoError(t, app.Register(ptr))

	var resolvedPtr *aService
	require.NoError(t, app.Exec(func(a *aService) {
		resolvedPtr = a
	}))
	require.NotEmpty(t, resolvedPtr)
	require.Equal(t, ptr, resolvedPtr)
	require.Equal(t, unsafe.Pointer(ptr), unsafe.Pointer(resolvedPtr))
}

func TestDI_MultiDependency(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.MultiRegister(
		func() *aService {
			return &aService{n: 100}
		},
		func(a *aService) bService {
			return bService{a: a}
		},
		func(a *aService, b bService) cService {
			return cService{a: a, b: b}
		},
	))

	require.NoError(t, app.Exec(func(a *aService, c cService) {
		require.NotNil(t, a)
		require.Equal(t, 100, a.n)
		require.Equal(t, 100, c.b.a.n)
	}))
}

func TestDI_ProvideFunctionFactory(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() func(*aService) *aService {
		return func(s *aService) *aService {
			ss := *s
			ss.n++
			return &ss
		}
	}))

	a := &aService{5}
	require.NoError(t, app.Exec(func(add func(*aService) *aService) {
		a = add(a)
	}))
	require.Equal(t, 6, a.n)
}

func TestDI_CircularDependencyProtection(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func(bService) *aService {
		return nil
	}))

	require.NoError(t, app.Register(func(*aService) bService {
		return bService{}
	}))

	err := app.Exec(func(bService) {})
	require.Equal(t, "providing type di.bService: providing type *di.aService: providing type di.bService: circular dependency detected", err.Error())
}

func TestDI_DuplicateRegistration(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return nil
	}))

	err := app.Register(func() *aService {
		return nil
	})
	require.Equal(t, "provider for type *di.aService already registered", err.Error())
}

func TestDI_StarterDependencyIsStartedBeforeProviding(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func(lc *Lifecycle) *aService {
		a := &aService{}
		lc.RegisterStarter(func(_ chan<- error, _ <-chan struct{}) error {
			a.n = 100
			return nil
		})
		return a
	}))

	require.NoError(t, app.Exec(func(a *aService) {
		require.Equal(t, 100, a.n)
	}))
}

func TestDI_LazyDepEvaluation(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		require.Fail(t, "should not be called")
		return nil
	}))

	require.NoError(t, app.Exec(func() {}))
}

func TestDI_MultipleInterfacesSatisfaction(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() *aService {
		return &aService{n: 10}
	}))

	require.NoError(t, app.Exec(func(n SmallNum, nn BigNum) {
		require.Equal(t, 10, n.Num())
		require.Equal(t, 20, nn.Num2())
	}))
}

func TestDI_ConversionBetweenInterfaces(t *testing.T) {
	app := newTestApplication()

	require.NoError(t, app.Register(func() BigNum {
		return &aService{n: 10}
	}))

	require.NoError(t, app.Exec(func(n SmallNum, nn BigNum) {
		require.Equal(t, 10, n.Num())
		require.Equal(t, 20, nn.Num2())
	}))
}

func TestDI_WaitForTerminationAndClose(t *testing.T) {
	triggerShutdownCh := make(chan struct{})

	app, err := NewApplication(Params{
		Logger:  pkgtest.Logger,
		CloseCh: triggerShutdownCh,
	})
	require.NoError(t, err)

	var (
		aStopped bool
		bStopped bool
	)

	require.NoError(t, app.MultiRegister(
		func(lc *Lifecycle) *aService {
			lc.RegisterStopper(func(_ context.Context) error {
				aStopped = true
				return nil
			})

			return &aService{}
		},
		func(lc *Lifecycle) bService {
			lc.RegisterStopper(func(_ context.Context) error {
				bStopped = true
				return nil
			})

			return bService{}
		},
	))

	require.NoError(t, app.Exec(func(*aService, bService) {}))

	closed := make(chan struct{})
	go func() {
		app.WaitForTerminationAndClose()
		close(closed)
	}()

	close(triggerShutdownCh)

	select {
	case <-closed:
	case <-time.After(1 * time.Second):
		require.Fail(t, "application did not close")
	}

	require.True(t, aStopped)
	require.True(t, bStopped)
}

func TestDI_Lifecycle(t *testing.T) {
	app := newTestApplication()

	a := &aService{n: 0}
	require.NoError(t, app.Register(func(lf *Lifecycle) *aService {
		lf.RegisterStarter(func(_ chan<- error, _ <-chan struct{}) error {
			a.n += 1
			return nil
		})

		lf.RegisterStopper(func(_ context.Context) error {
			a.n += 5
			return nil
		})

		return a
	}))

	require.NoError(t, app.Exec(func(a *aService) {}))

	app.Close()

	require.Equal(t, 6, a.n)
}

func TestDI_ApplicationTerminatesOnceErrorIsSent(t *testing.T) {
	app := newTestApplication()

	var (
		startedCh      = make(chan struct{})
		forwardCh      = make(chan struct{})
		appClosedAckCh = make(chan struct{})
	)
	require.NoError(t, app.MultiRegister(
		func(lc *Lifecycle) *aService {
			lc.RegisterStarter(func(errCh chan<- error, appClosedCh <-chan struct{}) error {
				go func() {
					<-forwardCh
					errCh <- fmt.Errorf("broken")

					<-appClosedCh
					close(appClosedAckCh)

				}()

				close(startedCh)
				return nil
			})

			return &aService{}
		}),
	)

	closed := make(chan struct{})
	go func() {
		require.NoError(t, app.Exec(func(_ *aService) {}))
		app.WaitForTerminationAndClose()
		close(closed)
	}()

	select {
	case <-startedCh:
	case <-time.After(1 * time.Second):
		require.Fail(t, "service not started")
	}

	close(forwardCh)

	select {
	case <-appClosedAckCh:
	case <-time.After(1 * time.Second):
		require.Fail(t, "app did not notified about closing")
	}

	select {
	case <-closed:
	case <-time.After(1 * time.Second):
		require.Fail(t, "application did not close")
	}
}

func newTestApplication() *Application {
	app, _ := NewApplication(Params{
		Logger: pkgtest.Logger,
	})
	return app
}

type cService struct {
	a *aService
	b bService
}

type bService struct {
	a *aService
}

type aService struct {
	n int
}

func (a *aService) String() string {
	return fmt.Sprintf("%d", a.n)
}

func (a *aService) Num() int {
	return a.n
}

func (a *aService) Num2() int {
	return a.n * 2
}

type SmallNum interface {
	Num() int
}

type BigNum interface {
	Num() int
	Num2() int
}
