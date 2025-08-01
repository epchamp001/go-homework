// Code generated by http://github.com/gojuno/minimock (v3.4.5). DO NOT EDIT.

package mock

//go:generate minimock -i pvz-cli/internal/usecase/packaging.PackagingStrategy -o packaging_strategy_mock.go -n PackagingStrategyMock -p mock

import (
	"pvz-cli/internal/domain/models"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// PackagingStrategyMock implements mm_packaging.PackagingStrategy
type PackagingStrategyMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcSurcharge          func() (p1 models.PriceKopecks)
	funcSurchargeOrigin    string
	inspectFuncSurcharge   func()
	afterSurchargeCounter  uint64
	beforeSurchargeCounter uint64
	SurchargeMock          mPackagingStrategyMockSurcharge

	funcValidate          func(weight float64) (err error)
	funcValidateOrigin    string
	inspectFuncValidate   func(weight float64)
	afterValidateCounter  uint64
	beforeValidateCounter uint64
	ValidateMock          mPackagingStrategyMockValidate
}

// NewPackagingStrategyMock returns a mock for mm_packaging.PackagingStrategy
func NewPackagingStrategyMock(t minimock.Tester) *PackagingStrategyMock {
	m := &PackagingStrategyMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.SurchargeMock = mPackagingStrategyMockSurcharge{mock: m}

	m.ValidateMock = mPackagingStrategyMockValidate{mock: m}
	m.ValidateMock.callArgs = []*PackagingStrategyMockValidateParams{}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mPackagingStrategyMockSurcharge struct {
	optional           bool
	mock               *PackagingStrategyMock
	defaultExpectation *PackagingStrategyMockSurchargeExpectation
	expectations       []*PackagingStrategyMockSurchargeExpectation

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// PackagingStrategyMockSurchargeExpectation specifies expectation struct of the PackagingStrategy.Surcharge
type PackagingStrategyMockSurchargeExpectation struct {
	mock *PackagingStrategyMock

	results      *PackagingStrategyMockSurchargeResults
	returnOrigin string
	Counter      uint64
}

// PackagingStrategyMockSurchargeResults contains results of the PackagingStrategy.Surcharge
type PackagingStrategyMockSurchargeResults struct {
	p1 models.PriceKopecks
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmSurcharge *mPackagingStrategyMockSurcharge) Optional() *mPackagingStrategyMockSurcharge {
	mmSurcharge.optional = true
	return mmSurcharge
}

// Expect sets up expected params for PackagingStrategy.Surcharge
func (mmSurcharge *mPackagingStrategyMockSurcharge) Expect() *mPackagingStrategyMockSurcharge {
	if mmSurcharge.mock.funcSurcharge != nil {
		mmSurcharge.mock.t.Fatalf("PackagingStrategyMock.Surcharge mock is already set by Set")
	}

	if mmSurcharge.defaultExpectation == nil {
		mmSurcharge.defaultExpectation = &PackagingStrategyMockSurchargeExpectation{}
	}

	return mmSurcharge
}

// Inspect accepts an inspector function that has same arguments as the PackagingStrategy.Surcharge
func (mmSurcharge *mPackagingStrategyMockSurcharge) Inspect(f func()) *mPackagingStrategyMockSurcharge {
	if mmSurcharge.mock.inspectFuncSurcharge != nil {
		mmSurcharge.mock.t.Fatalf("Inspect function is already set for PackagingStrategyMock.Surcharge")
	}

	mmSurcharge.mock.inspectFuncSurcharge = f

	return mmSurcharge
}

// Return sets up results that will be returned by PackagingStrategy.Surcharge
func (mmSurcharge *mPackagingStrategyMockSurcharge) Return(p1 models.PriceKopecks) *PackagingStrategyMock {
	if mmSurcharge.mock.funcSurcharge != nil {
		mmSurcharge.mock.t.Fatalf("PackagingStrategyMock.Surcharge mock is already set by Set")
	}

	if mmSurcharge.defaultExpectation == nil {
		mmSurcharge.defaultExpectation = &PackagingStrategyMockSurchargeExpectation{mock: mmSurcharge.mock}
	}
	mmSurcharge.defaultExpectation.results = &PackagingStrategyMockSurchargeResults{p1}
	mmSurcharge.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmSurcharge.mock
}

// Set uses given function f to mock the PackagingStrategy.Surcharge method
func (mmSurcharge *mPackagingStrategyMockSurcharge) Set(f func() (p1 models.PriceKopecks)) *PackagingStrategyMock {
	if mmSurcharge.defaultExpectation != nil {
		mmSurcharge.mock.t.Fatalf("Default expectation is already set for the PackagingStrategy.Surcharge method")
	}

	if len(mmSurcharge.expectations) > 0 {
		mmSurcharge.mock.t.Fatalf("Some expectations are already set for the PackagingStrategy.Surcharge method")
	}

	mmSurcharge.mock.funcSurcharge = f
	mmSurcharge.mock.funcSurchargeOrigin = minimock.CallerInfo(1)
	return mmSurcharge.mock
}

// Times sets number of times PackagingStrategy.Surcharge should be invoked
func (mmSurcharge *mPackagingStrategyMockSurcharge) Times(n uint64) *mPackagingStrategyMockSurcharge {
	if n == 0 {
		mmSurcharge.mock.t.Fatalf("Times of PackagingStrategyMock.Surcharge mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmSurcharge.expectedInvocations, n)
	mmSurcharge.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmSurcharge
}

func (mmSurcharge *mPackagingStrategyMockSurcharge) invocationsDone() bool {
	if len(mmSurcharge.expectations) == 0 && mmSurcharge.defaultExpectation == nil && mmSurcharge.mock.funcSurcharge == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmSurcharge.mock.afterSurchargeCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmSurcharge.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Surcharge implements mm_packaging.PackagingStrategy
func (mmSurcharge *PackagingStrategyMock) Surcharge() (p1 models.PriceKopecks) {
	mm_atomic.AddUint64(&mmSurcharge.beforeSurchargeCounter, 1)
	defer mm_atomic.AddUint64(&mmSurcharge.afterSurchargeCounter, 1)

	mmSurcharge.t.Helper()

	if mmSurcharge.inspectFuncSurcharge != nil {
		mmSurcharge.inspectFuncSurcharge()
	}

	if mmSurcharge.SurchargeMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmSurcharge.SurchargeMock.defaultExpectation.Counter, 1)

		mm_results := mmSurcharge.SurchargeMock.defaultExpectation.results
		if mm_results == nil {
			mmSurcharge.t.Fatal("No results are set for the PackagingStrategyMock.Surcharge")
		}
		return (*mm_results).p1
	}
	if mmSurcharge.funcSurcharge != nil {
		return mmSurcharge.funcSurcharge()
	}
	mmSurcharge.t.Fatalf("Unexpected call to PackagingStrategyMock.Surcharge.")
	return
}

// SurchargeAfterCounter returns a count of finished PackagingStrategyMock.Surcharge invocations
func (mmSurcharge *PackagingStrategyMock) SurchargeAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSurcharge.afterSurchargeCounter)
}

// SurchargeBeforeCounter returns a count of PackagingStrategyMock.Surcharge invocations
func (mmSurcharge *PackagingStrategyMock) SurchargeBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSurcharge.beforeSurchargeCounter)
}

// MinimockSurchargeDone returns true if the count of the Surcharge invocations corresponds
// the number of defined expectations
func (m *PackagingStrategyMock) MinimockSurchargeDone() bool {
	if m.SurchargeMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.SurchargeMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.SurchargeMock.invocationsDone()
}

// MinimockSurchargeInspect logs each unmet expectation
func (m *PackagingStrategyMock) MinimockSurchargeInspect() {
	for _, e := range m.SurchargeMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to PackagingStrategyMock.Surcharge")
		}
	}

	afterSurchargeCounter := mm_atomic.LoadUint64(&m.afterSurchargeCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.SurchargeMock.defaultExpectation != nil && afterSurchargeCounter < 1 {
		m.t.Errorf("Expected call to PackagingStrategyMock.Surcharge at\n%s", m.SurchargeMock.defaultExpectation.returnOrigin)
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSurcharge != nil && afterSurchargeCounter < 1 {
		m.t.Errorf("Expected call to PackagingStrategyMock.Surcharge at\n%s", m.funcSurchargeOrigin)
	}

	if !m.SurchargeMock.invocationsDone() && afterSurchargeCounter > 0 {
		m.t.Errorf("Expected %d calls to PackagingStrategyMock.Surcharge at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.SurchargeMock.expectedInvocations), m.SurchargeMock.expectedInvocationsOrigin, afterSurchargeCounter)
	}
}

type mPackagingStrategyMockValidate struct {
	optional           bool
	mock               *PackagingStrategyMock
	defaultExpectation *PackagingStrategyMockValidateExpectation
	expectations       []*PackagingStrategyMockValidateExpectation

	callArgs []*PackagingStrategyMockValidateParams
	mutex    sync.RWMutex

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// PackagingStrategyMockValidateExpectation specifies expectation struct of the PackagingStrategy.Validate
type PackagingStrategyMockValidateExpectation struct {
	mock               *PackagingStrategyMock
	params             *PackagingStrategyMockValidateParams
	paramPtrs          *PackagingStrategyMockValidateParamPtrs
	expectationOrigins PackagingStrategyMockValidateExpectationOrigins
	results            *PackagingStrategyMockValidateResults
	returnOrigin       string
	Counter            uint64
}

// PackagingStrategyMockValidateParams contains parameters of the PackagingStrategy.Validate
type PackagingStrategyMockValidateParams struct {
	weight float64
}

// PackagingStrategyMockValidateParamPtrs contains pointers to parameters of the PackagingStrategy.Validate
type PackagingStrategyMockValidateParamPtrs struct {
	weight *float64
}

// PackagingStrategyMockValidateResults contains results of the PackagingStrategy.Validate
type PackagingStrategyMockValidateResults struct {
	err error
}

// PackagingStrategyMockValidateOrigins contains origins of expectations of the PackagingStrategy.Validate
type PackagingStrategyMockValidateExpectationOrigins struct {
	origin       string
	originWeight string
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmValidate *mPackagingStrategyMockValidate) Optional() *mPackagingStrategyMockValidate {
	mmValidate.optional = true
	return mmValidate
}

// Expect sets up expected params for PackagingStrategy.Validate
func (mmValidate *mPackagingStrategyMockValidate) Expect(weight float64) *mPackagingStrategyMockValidate {
	if mmValidate.mock.funcValidate != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by Set")
	}

	if mmValidate.defaultExpectation == nil {
		mmValidate.defaultExpectation = &PackagingStrategyMockValidateExpectation{}
	}

	if mmValidate.defaultExpectation.paramPtrs != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by ExpectParams functions")
	}

	mmValidate.defaultExpectation.params = &PackagingStrategyMockValidateParams{weight}
	mmValidate.defaultExpectation.expectationOrigins.origin = minimock.CallerInfo(1)
	for _, e := range mmValidate.expectations {
		if minimock.Equal(e.params, mmValidate.defaultExpectation.params) {
			mmValidate.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmValidate.defaultExpectation.params)
		}
	}

	return mmValidate
}

// ExpectWeightParam1 sets up expected param weight for PackagingStrategy.Validate
func (mmValidate *mPackagingStrategyMockValidate) ExpectWeightParam1(weight float64) *mPackagingStrategyMockValidate {
	if mmValidate.mock.funcValidate != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by Set")
	}

	if mmValidate.defaultExpectation == nil {
		mmValidate.defaultExpectation = &PackagingStrategyMockValidateExpectation{}
	}

	if mmValidate.defaultExpectation.params != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by Expect")
	}

	if mmValidate.defaultExpectation.paramPtrs == nil {
		mmValidate.defaultExpectation.paramPtrs = &PackagingStrategyMockValidateParamPtrs{}
	}
	mmValidate.defaultExpectation.paramPtrs.weight = &weight
	mmValidate.defaultExpectation.expectationOrigins.originWeight = minimock.CallerInfo(1)

	return mmValidate
}

// Inspect accepts an inspector function that has same arguments as the PackagingStrategy.Validate
func (mmValidate *mPackagingStrategyMockValidate) Inspect(f func(weight float64)) *mPackagingStrategyMockValidate {
	if mmValidate.mock.inspectFuncValidate != nil {
		mmValidate.mock.t.Fatalf("Inspect function is already set for PackagingStrategyMock.Validate")
	}

	mmValidate.mock.inspectFuncValidate = f

	return mmValidate
}

// Return sets up results that will be returned by PackagingStrategy.Validate
func (mmValidate *mPackagingStrategyMockValidate) Return(err error) *PackagingStrategyMock {
	if mmValidate.mock.funcValidate != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by Set")
	}

	if mmValidate.defaultExpectation == nil {
		mmValidate.defaultExpectation = &PackagingStrategyMockValidateExpectation{mock: mmValidate.mock}
	}
	mmValidate.defaultExpectation.results = &PackagingStrategyMockValidateResults{err}
	mmValidate.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmValidate.mock
}

// Set uses given function f to mock the PackagingStrategy.Validate method
func (mmValidate *mPackagingStrategyMockValidate) Set(f func(weight float64) (err error)) *PackagingStrategyMock {
	if mmValidate.defaultExpectation != nil {
		mmValidate.mock.t.Fatalf("Default expectation is already set for the PackagingStrategy.Validate method")
	}

	if len(mmValidate.expectations) > 0 {
		mmValidate.mock.t.Fatalf("Some expectations are already set for the PackagingStrategy.Validate method")
	}

	mmValidate.mock.funcValidate = f
	mmValidate.mock.funcValidateOrigin = minimock.CallerInfo(1)
	return mmValidate.mock
}

// When sets expectation for the PackagingStrategy.Validate which will trigger the result defined by the following
// Then helper
func (mmValidate *mPackagingStrategyMockValidate) When(weight float64) *PackagingStrategyMockValidateExpectation {
	if mmValidate.mock.funcValidate != nil {
		mmValidate.mock.t.Fatalf("PackagingStrategyMock.Validate mock is already set by Set")
	}

	expectation := &PackagingStrategyMockValidateExpectation{
		mock:               mmValidate.mock,
		params:             &PackagingStrategyMockValidateParams{weight},
		expectationOrigins: PackagingStrategyMockValidateExpectationOrigins{origin: minimock.CallerInfo(1)},
	}
	mmValidate.expectations = append(mmValidate.expectations, expectation)
	return expectation
}

// Then sets up PackagingStrategy.Validate return parameters for the expectation previously defined by the When method
func (e *PackagingStrategyMockValidateExpectation) Then(err error) *PackagingStrategyMock {
	e.results = &PackagingStrategyMockValidateResults{err}
	return e.mock
}

// Times sets number of times PackagingStrategy.Validate should be invoked
func (mmValidate *mPackagingStrategyMockValidate) Times(n uint64) *mPackagingStrategyMockValidate {
	if n == 0 {
		mmValidate.mock.t.Fatalf("Times of PackagingStrategyMock.Validate mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmValidate.expectedInvocations, n)
	mmValidate.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmValidate
}

func (mmValidate *mPackagingStrategyMockValidate) invocationsDone() bool {
	if len(mmValidate.expectations) == 0 && mmValidate.defaultExpectation == nil && mmValidate.mock.funcValidate == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmValidate.mock.afterValidateCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmValidate.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Validate implements mm_packaging.PackagingStrategy
func (mmValidate *PackagingStrategyMock) Validate(weight float64) (err error) {
	mm_atomic.AddUint64(&mmValidate.beforeValidateCounter, 1)
	defer mm_atomic.AddUint64(&mmValidate.afterValidateCounter, 1)

	mmValidate.t.Helper()

	if mmValidate.inspectFuncValidate != nil {
		mmValidate.inspectFuncValidate(weight)
	}

	mm_params := PackagingStrategyMockValidateParams{weight}

	// Record call args
	mmValidate.ValidateMock.mutex.Lock()
	mmValidate.ValidateMock.callArgs = append(mmValidate.ValidateMock.callArgs, &mm_params)
	mmValidate.ValidateMock.mutex.Unlock()

	for _, e := range mmValidate.ValidateMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmValidate.ValidateMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmValidate.ValidateMock.defaultExpectation.Counter, 1)
		mm_want := mmValidate.ValidateMock.defaultExpectation.params
		mm_want_ptrs := mmValidate.ValidateMock.defaultExpectation.paramPtrs

		mm_got := PackagingStrategyMockValidateParams{weight}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.weight != nil && !minimock.Equal(*mm_want_ptrs.weight, mm_got.weight) {
				mmValidate.t.Errorf("PackagingStrategyMock.Validate got unexpected parameter weight, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
					mmValidate.ValidateMock.defaultExpectation.expectationOrigins.originWeight, *mm_want_ptrs.weight, mm_got.weight, minimock.Diff(*mm_want_ptrs.weight, mm_got.weight))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmValidate.t.Errorf("PackagingStrategyMock.Validate got unexpected parameters, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
				mmValidate.ValidateMock.defaultExpectation.expectationOrigins.origin, *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmValidate.ValidateMock.defaultExpectation.results
		if mm_results == nil {
			mmValidate.t.Fatal("No results are set for the PackagingStrategyMock.Validate")
		}
		return (*mm_results).err
	}
	if mmValidate.funcValidate != nil {
		return mmValidate.funcValidate(weight)
	}
	mmValidate.t.Fatalf("Unexpected call to PackagingStrategyMock.Validate. %v", weight)
	return
}

// ValidateAfterCounter returns a count of finished PackagingStrategyMock.Validate invocations
func (mmValidate *PackagingStrategyMock) ValidateAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmValidate.afterValidateCounter)
}

// ValidateBeforeCounter returns a count of PackagingStrategyMock.Validate invocations
func (mmValidate *PackagingStrategyMock) ValidateBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmValidate.beforeValidateCounter)
}

// Calls returns a list of arguments used in each call to PackagingStrategyMock.Validate.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmValidate *mPackagingStrategyMockValidate) Calls() []*PackagingStrategyMockValidateParams {
	mmValidate.mutex.RLock()

	argCopy := make([]*PackagingStrategyMockValidateParams, len(mmValidate.callArgs))
	copy(argCopy, mmValidate.callArgs)

	mmValidate.mutex.RUnlock()

	return argCopy
}

// MinimockValidateDone returns true if the count of the Validate invocations corresponds
// the number of defined expectations
func (m *PackagingStrategyMock) MinimockValidateDone() bool {
	if m.ValidateMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.ValidateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.ValidateMock.invocationsDone()
}

// MinimockValidateInspect logs each unmet expectation
func (m *PackagingStrategyMock) MinimockValidateInspect() {
	for _, e := range m.ValidateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to PackagingStrategyMock.Validate at\n%s with params: %#v", e.expectationOrigins.origin, *e.params)
		}
	}

	afterValidateCounter := mm_atomic.LoadUint64(&m.afterValidateCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.ValidateMock.defaultExpectation != nil && afterValidateCounter < 1 {
		if m.ValidateMock.defaultExpectation.params == nil {
			m.t.Errorf("Expected call to PackagingStrategyMock.Validate at\n%s", m.ValidateMock.defaultExpectation.returnOrigin)
		} else {
			m.t.Errorf("Expected call to PackagingStrategyMock.Validate at\n%s with params: %#v", m.ValidateMock.defaultExpectation.expectationOrigins.origin, *m.ValidateMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcValidate != nil && afterValidateCounter < 1 {
		m.t.Errorf("Expected call to PackagingStrategyMock.Validate at\n%s", m.funcValidateOrigin)
	}

	if !m.ValidateMock.invocationsDone() && afterValidateCounter > 0 {
		m.t.Errorf("Expected %d calls to PackagingStrategyMock.Validate at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.ValidateMock.expectedInvocations), m.ValidateMock.expectedInvocationsOrigin, afterValidateCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *PackagingStrategyMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockSurchargeInspect()

			m.MinimockValidateInspect()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *PackagingStrategyMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *PackagingStrategyMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockSurchargeDone() &&
		m.MinimockValidateDone()
}
