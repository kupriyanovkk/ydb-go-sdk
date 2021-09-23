package table

import (
	"container/list"
	"context"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/deadline"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
	"sync"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"

	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/errors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/timeutil"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
)

var (
	DefaultSessionPoolKeepAliveTimeout     = 500 * time.Millisecond
	DefaultSessionPoolDeleteTimeout        = 500 * time.Millisecond
	DefaultSessionPoolCreateSessionTimeout = 5 * time.Second
	DefaultSessionPoolIdleThreshold        = 5 * time.Minute
	DefaultSessionPoolSizeLimit            = 50
	DefaultKeepAliveMinSize                = 10
	DefaultIdleKeepAliveThreshold          = 2
)

var (
	// ErrSessionPoolClosed is returned by a pool instance to indicate
	// that pool is closed and not able to complete requested operation.
	ErrSessionPoolClosed = errors.New("ydb: table: session pool is closed")

	// ErrSessionPoolOverflow is returned by a pool instance to indicate
	// that the pool is full and requested operation is not able to complete.
	ErrSessionPoolOverflow = errors.New("ydb: table: session pool overflow")

	// ErrSessionUnknown is returned by a pool instance to indicate that
	// requested session does not exist within the pool.
	ErrSessionUnknown = errors.New("ydb: table: unknown session")

	// ErrNoProgress is returned by a pool instance to indicate that
	// operation could not be completed.
	ErrNoProgress = errors.New("ydb: table: no progress")
)

// SessionBuilder is the interface that holds logic of creating or deleting
// sessions.
type SessionBuilder interface {
	CreateSession(context.Context) (table.Session, error)
}

type Pool interface {
	Take(ctx context.Context, s table.Session) (took bool, err error)
	Put(ctx context.Context, s table.Session) (err error)
	Create(ctx context.Context) (s table.Session, err error)
	Retry(ctx context.Context, idempotent bool, op table.RetryOperation) (error, []error)
	Close(ctx context.Context) error
}

type ClientAsPool interface {
	table.Client
	Pool
}

// pool is a set of session instances that may be reused.
// A pool is safe for use by multiple goroutines simultaneously.
type pool struct {
	// Trace is an optional session lifetime tracing options.
	Trace trace.Table

	// Builder holds an object capable for creating and deleting sessions.
	// It must not be nil.
	Builder SessionBuilder

	// SizeLimit is an upper bound of pooled sessions.
	// If SizeLimit is less than or equal to zero then the
	// DefaultSessionPoolSizeLimit variable is used as a limit.
	SizeLimit int

	// KeepAliveMinSize is a lower bound for sessions in the pool. If there are more sessions open, then
	// the excess idle ones will be closed and removed after IdleKeepAliveThreshold is reached for each of them.
	// If KeepAliveMinSize is less than zero, then no sessions will be preserved
	// If KeepAliveMinSize is zero, the DefaultKeepAliveMinSize is used
	KeepAliveMinSize int

	// IdleKeepAliveThreshold is a number of keepAlive messages to call before the
	// session is removed if it is an excess session (see KeepAliveMinSize)
	// This means that session will be deleted after the expiration of lifetime = IdleThreshold * IdleKeepAliveThreshold
	// If IdleKeepAliveThreshold is less than zero then it will be treated as infinite and no sessions will
	// be removed ever.
	// If IdleKeepAliveThreshold is equal to zero, it will be set to DefaultIdleKeepAliveThreshold
	IdleKeepAliveThreshold int

	// IdleLimit is an upper bound of pooled sessions without any activity
	// within.
	// IdleLimit int

	// IdleThreshold is a maximum duration between any activity within session.
	// If this threshold reached, KeepAlive() method will be called on idle
	// session.
	//
	// If IdleThreshold is less than zero then there is no idle limit.
	// If IdleThreshold is zero, then the DefaultSessionPoolIdleThreshold value
	// is used.
	IdleThreshold time.Duration

	// KeepAliveTimeout limits maximum time spent on KeepAlive request
	// If KeepAliveTimeout is less than or equal to zero then the
	// DefaultSessionPoolKeepAliveTimeout is used.
	KeepAliveTimeout time.Duration

	// CreateSessionTimeout limits maximum time spent on Create session request
	// If CreateSessionTimeout is less than or equal to zero then the
	// DefaultSessionPoolCreateSessionTimeout is used.
	CreateSessionTimeout time.Duration

	// DeleteTimeout limits maximum time spent on Delete request
	// If DeleteTimeout is less than or equal to zero then the
	// DefaultSessionPoolDeleteTimeout is used.
	DeleteTimeout time.Duration

	mu               sync.Mutex
	initOnce         sync.Once
	index            map[table.Session]sessionInfo
	createInProgress int        // KIKIMR-9163: in-create-process counter
	limit            int        // Upper bound for pool size.
	idle             *list.List // list<table.session>
	waitq            *list.List // list<*chan table.session>

	keeperWake chan struct{} // Set by keeper.
	keeperStop chan struct{}
	keeperDone chan struct{}

	touching     bool
	touchingDone chan struct{}

	closed bool

	waitChPool        sync.Pool
	testHookGetWaitCh func() // nil except some tests.
}

func (p *pool) isClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

func (p *pool) init() {
	p.initOnce.Do(func() {
		p.index = make(map[table.Session]sessionInfo)

		p.idle = list.New()
		p.waitq = list.New()
		p.limit = p.SizeLimit
		if p.limit <= 0 {
			p.limit = DefaultSessionPoolSizeLimit
		}

		if p.IdleThreshold == 0 {
			p.IdleThreshold = DefaultSessionPoolIdleThreshold
		}
		if p.IdleThreshold > 0 {
			p.keeperStop = make(chan struct{})
			p.keeperDone = make(chan struct{})
			go p.keeper()
		}

		if p.KeepAliveMinSize < 0 {
			p.KeepAliveMinSize = 0
		} else if p.KeepAliveMinSize == 0 {
			p.KeepAliveMinSize = DefaultKeepAliveMinSize
		}

		if p.IdleKeepAliveThreshold == 0 {
			p.IdleKeepAliveThreshold = DefaultIdleKeepAliveThreshold
		}

		if p.CreateSessionTimeout <= 0 {
			p.CreateSessionTimeout = DefaultSessionPoolCreateSessionTimeout
		}
		if p.DeleteTimeout <= 0 {
			p.DeleteTimeout = DefaultSessionPoolDeleteTimeout
		}
		if p.KeepAliveTimeout <= 0 {
			p.KeepAliveTimeout = DefaultSessionPoolKeepAliveTimeout
		}
	})
}

func isCreateSessionErrorRetriable(err error) bool {
	switch {
	case
		errors.Is(err, ErrSessionPoolOverflow),
		errors.IsOpError(err, errors.StatusOverloaded),
		errors.IsTransportError(err, errors.TransportErrorResourceExhausted),
		errors.IsTransportError(err, errors.TransportErrorDeadlineExceeded),
		errors.IsTransportError(err, errors.TransportErrorUnavailable):
		return true
	default:
		return false
	}
}

type createSessionResult struct {
	s   table.Session
	err error
}

// p.mu must NOT be held.
func (p *pool) createSession(ctx context.Context) (table.Session, error) {
	// pre-check the pool size
	p.mu.Lock()
	enoughSpace := p.createInProgress+len(p.index) < p.limit
	if enoughSpace {
		p.createInProgress++
	}
	p.mu.Unlock()

	if !enoughSpace {
		return nil, ErrSessionPoolOverflow
	}

	resCh := make(chan createSessionResult, 1) // for non-block write

	go func() {
		var r createSessionResult

		ctx, cancel := context.WithTimeout(
			deadline.ContextWithoutDeadline(ctx),
			p.CreateSessionTimeout,
		)

		defer func() {
			cancel()
			close(resCh)
		}()

		r.s, r.err = p.Builder.CreateSession(ctx)
		// if session not nil - error must be nil and vice versa
		if r.s == nil && r.err == nil {
			panic("ydb: abnormal result of pool.Builder.CreateSession()")
		}

		if r.err != nil {
			p.mu.Lock()
			p.createInProgress--
			p.mu.Unlock()
			resCh <- r
			return
		}
		r.s.OnClose(func() {
			p.mu.Lock()
			defer p.mu.Unlock()
			info, has := p.index[r.s]
			if !has {
				return
			}

			delete(p.index, r.s)
			p.notify(nil)

			if info.idle != nil {
				panic("ydb: table: session closed while still in idle pool")
			}
		})

		// Slot for session already reserved early
		p.mu.Lock()
		p.index[r.s] = sessionInfo{}
		p.createInProgress--
		p.mu.Unlock()

		resCh <- r
	}()

	select {
	case r := <-resCh:
		if r.s == nil && r.err == nil {
			panic("ydb: abnormal result of pool.createSession()")
		}
		return r.s, r.err
	case <-ctx.Done():
		// read result from resCh for prevention of forgetting session
		go func() {
			if r, ok := <-resCh; ok && r.s != nil {
				_ = r.s.Close(deadline.ContextWithoutDeadline(ctx))
			}
		}()
		return nil, ctx.Err()
	}
	return nil, nil
}

// Get returns first idle session from the pool and removes it from
// there. If no items stored in pool it creates new one by calling
// Builder.CreateSession() method and returns it.
func (p *pool) Get(ctx context.Context) (s table.Session, err error) {
	p.init()

	var (
		i     = 0
		start = time.Now()
	)
	getDone := trace.TablePoolOnGet(p.Trace.TablePool, ctx)
	defer func() {
		getDone(ctx, s.ID(), time.Since(start), i, err)
	}()

	const maxAttempts = 100
	for ; s == nil && err == nil && i < maxAttempts; i++ {
		var (
			ch *chan table.Session
			el *list.Element // Element in the wait queue.
		)

		if p.isClosed() {
			return nil, ErrSessionPoolClosed
		}
		p.mu.Lock()
		s = p.removeFirstIdle()
		p.mu.Unlock()

		if s == nil {
			// Try create new session without awaiting for reused one.
			s, err = p.createSession(ctx)
			// got session or err is not recoverable
			if s != nil || err != nil && !isCreateSessionErrorRetriable(err) {
				return s, err
			}
			err = nil
		}

		// get here after check isCreateSessionErrorRetriable
		if s == nil {
			// Try to wait for a touched session - pool is full.
			//
			// This should be done only if number of currently waiting goroutines
			// are less than maximum amount of touched session. That is, we want to
			// be fair here and not to lock more goroutines than we could ship
			// session to.
			p.mu.Lock()
			s = p.removeFirstIdle()
			if s != nil {
				p.mu.Unlock()
				return s, nil
			}
			ch = p.getWaitCh()
			el = p.waitq.PushBack(ch)
			p.mu.Unlock()
		}

		if ch == nil {
			continue
		}
		waitDone := trace.TablePoolOnWait(p.Trace.TablePool, ctx)
		var ok bool
		select {
		case s, ok = <-*ch:
			// Note that race may occur and some goroutine may try to write
			// session into channel after it was enqueued but before it being
			// read here. In that case we will receive nil here and will retry.
			//
			// The same path will work when some session become deleted - the
			// nil value will be sent into the channel.
			if ok {
				// Put only filled and not closed channel back to the pool.
				// That is, we need to avoid races on filling reused channel
				// for the next waiter – session could be lost for a long time.
				p.putWaitCh(ch)
			}
			waitDone(ctx, s.ID(), err)

		case <-ctx.Done():
			p.mu.Lock()
			// Note that el can be already removed here while we were moving
			// from reading from ch to this case. This does not make any
			// difference – channel will be closed by notifying goroutine.
			p.waitq.Remove(el)
			p.mu.Unlock()
			err = ctx.Err()
			waitDone(ctx, s.ID(), err)
			return nil, err
		}
	}
	if s == nil && err == nil {
		err = ErrNoProgress
	}

	return s, err
}

// Put returns session to the pool for further reuse.
// If pool is already closed Put() calls s.Close(ctx) and returns
// ErrSessionPoolClosed.
// If pool is overflow calls s.Close(ctx) and returns
// ErrSessionPoolOverflow.
//
// Note that Put() must be called only once after being created or received by
// Get() or Take() calls. In other way it will produce unexpected behavior or
// panic.
func (p *pool) Put(ctx context.Context, s table.Session) (err error) {
	p.init()

	putDone := trace.TablePoolOnPut(p.Trace.TablePool, ctx, s.ID())
	defer func() {
		putDone(ctx, s.ID(), err)
	}()

	p.mu.Lock()
	switch {
	case p.closed:
		err = ErrSessionPoolClosed

	case p.idle.Len() >= p.limit:
		err = ErrSessionPoolOverflow

	default:
		if !p.notify(s) {
			p.pushIdle(s, timeutil.Now())
		}
	}
	p.mu.Unlock()

	if err != nil {
		_ = p.CloseSession(ctx, s)
	}

	return
}

// Take removes session s from the pool and ensures that s will not be returned
// by other Take() or Get() calls.
//
// The intended way of Take() use is to create session by calling Create() and
// Put() it later to prepare KeepAlive tracking when session is idle. When
// session becomes active, one should call Take() to stop KeepAlive tracking
// (simultaneous use of session is prohibited).
//
// After session returned to the pool by calling PutBusy() it can not be taken
// by Take() any more. That is, semantically PutBusy() is the same as session's
// Close().
//
// It is assumed that Take() callers never call Get() method.
func (p *pool) Take(ctx context.Context, s table.Session) (took bool, err error) {
	p.init()

	takeWait := trace.TablePoolOnTake(p.Trace.TablePool, ctx, s.ID())
	var takeDone func(_ context.Context, _ string, took bool, _ error)

	if p.isClosed() {
		return false, ErrSessionPoolClosed
	}
	var has bool
	p.mu.Lock()
	for has, took = p.takeIdle(s); has && !took && p.touching; has, took = p.takeIdle(s) {
		cond := p.touchCond()
		p.mu.Unlock()
		takeDone = takeWait(ctx, s.ID())

		// Keepalive processing takes place right now.
		// Try to await touched session before creation of new one.
		select {
		case <-cond:

		case <-ctx.Done():
			return false, ctx.Err()
		}

		p.mu.Lock()
	}
	defer func() {
		if takeDone != nil {
			takeDone(ctx, s.ID(), took, err)
		}
	}()
	p.mu.Unlock()

	if !has {
		err = ErrSessionUnknown
	}

	return took, err
}

// Create creates new session and returns it.
// The intended way of Create() usage relates to Take() method.
func (p *pool) Create(ctx context.Context) (s table.Session, err error) {
	p.init()

	createDone := trace.TablePoolOnCreate(p.Trace.TablePool, ctx)
	defer func() {
		createDone(ctx, s.ID(), err)
	}()

	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		p.mu.Lock()
		// NOTE: here is a race condition with keeper() running.
		// session could be deleted by some reason after we released the mutex.
		// We are not dealing with this because even if session is not deleted
		// by keeper() it could be staled on the server and the same user
		// experience will appear.
		s, _ = p.peekFirstIdle()
		p.mu.Unlock()

		if s == nil {
			return p.createSession(ctx)
		}

		took, err := p.Take(ctx, s)
		if err == nil && !took || errors.Is(err, ErrSessionUnknown) {
			// session was marked for deletion or deleted by keeper() - race happen - retry
			s = nil
			err = nil
			continue
		}
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	return nil, ErrNoProgress
}

// Close deletes all stored sessions inside pool.
// It also stops all underlying timers and goroutines.
// It returns first error occurred during stale sessions' deletion.
// Note that even on error it calls Close() on each session.
func (p *pool) Close(ctx context.Context) (err error) {
	p.init()

	closeDone := trace.TablePoolOnClose(p.Trace.TablePool, ctx)
	defer func() {
		closeDone(ctx, err)
	}()

	if p.isClosed() {
		return
	}

	p.mu.Lock()
	p.closed = true
	keeperDone := p.keeperDone
	if ch := p.keeperStop; ch != nil {
		close(ch)
	}
	p.mu.Unlock()

	if keeperDone != nil {
		<-keeperDone
	}

	p.mu.Lock()
	idle := p.idle
	waitq := p.waitq
	p.limit = 0
	p.idle = list.New()
	p.waitq = list.New()
	p.index = make(map[table.Session]sessionInfo)
	p.mu.Unlock()

	for el := waitq.Front(); el != nil; el = el.Next() {
		ch := el.Value.(*chan table.Session)
		close(*ch)
	}
	for e := idle.Front(); e != nil; e = e.Next() {
		s := e.Value.(table.Session)
		_ = p.CloseSession(ctx, s)
	}

	return nil
}

// Retry provide the best effort fo retrying operation
// Retry implements internal busy loop until one of the following conditions is met:
// - deadline was cancelled or deadlined
// - retry operation returned nil as error
// Warning: if deadline without deadline or cancellation func Retry will be worked infinite
func (p *pool) Retry(ctx context.Context, retryNoIdempotent bool, op table.RetryOperation) (err error, issues []error) {
	return retryBackoff(ctx, p, retry.FastBackoff, retry.SlowBackoff, retryNoIdempotent, op)
}

func (p *pool) Stats() poolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	idleCount, waitQCount, indexCount := 0, 0, 0
	if p.idle != nil {
		idleCount = p.idle.Len()
	}
	if p.waitq != nil {
		waitQCount = p.waitq.Len()
	}
	if p.index != nil {
		indexCount = len(p.index)
	}
	return poolStats{
		Idle:             idleCount,
		Index:            indexCount,
		WaitQ:            waitQCount,
		CreateInProgress: p.createInProgress,
		MinSize:          p.KeepAliveMinSize,
		MaxSize:          p.limit,
	}
}

func (p *pool) keeper() {
	defer close(p.keeperDone)
	var (
		toTouch    []table.Session // Cached for reuse.
		toDelete   []table.Session // Cached for reuse.
		toTryAgain []table.Session // Cached for reuse.

		wake  = make(chan struct{})
		timer = timeutil.NewTimer(p.IdleThreshold)
	)

	for {
		var now time.Time
		select {
		case <-wake:
			wake = make(chan struct{})
			if !timer.Stop() {
				select {
				case <-timer.C():
				default:
				}
			}
			timer.Reset(p.IdleThreshold)
			continue

		case <-p.keeperStop:
			return

		case now = <-timer.C():
			toTouch = toTouch[:0]
			toTryAgain = toTryAgain[:0]
			toDelete = toDelete[:0]

			p.mu.Lock()
			{
				p.touching = true
				for p.idle.Len() > 0 {
					s, touched := p.peekFirstIdle()
					if s == nil || now.Sub(touched) < p.IdleThreshold {
						break
					}
					_ = p.removeIdle(s)
					toTouch = append(toTouch, s)
				}
			}
			p.mu.Unlock()

			var mark *list.Element // Element in the list to insert touched sessions after.
			for i, s := range toTouch {
				toTouch[i] = nil

				p.mu.Lock()
				keepAliveCount := p.incrementKeepAlive(s)
				lenIndex := len(p.index)
				p.mu.Unlock()
				// if keepAlive was called more than the corresponding limit for the session to be alive and more
				// sessions are open than the lower limit of continuously kept sessions
				if p.IdleKeepAliveThreshold > 0 && keepAliveCount >= p.IdleKeepAliveThreshold &&
					p.KeepAliveMinSize < lenIndex-len(toDelete) {

					toDelete = append(toDelete, s)
					continue
				}

				_, err := p.keepAliveSession(context.Background(), s)
				if err != nil {
					switch {
					case
						errors.IsOpError(err, errors.StatusBadSession),
						errors.IsTransportError(err, errors.TransportErrorCanceled),
						errors.IsTransportError(err, errors.TransportErrorDeadlineExceeded):
						toDelete = append(toDelete, s)
					default:
						toTryAgain = append(toTryAgain, s)
					}
					continue
				}

				p.mu.Lock()
				if !p.notify(s) {
					// Need to push back session into list in order, to prevent
					// shuffling of sessions order.
					//
					// That is, there may be a race condition, when some session S1
					// pushed back in the list before we took the mutex. Suppose S1
					// touched time is greater than ours `now` for S0. If so, it
					// then may interrupt next keep alive iteration earlier and
					// prevent our session S0 being touched:
					// time.Since(S1) < threshold but time.Since(S0) > threshold.
					mark = p.pushIdleInOrderAfter(s, now, mark)
				}
				p.mu.Unlock()
			}

			{ // push all the soft failed sessions to retry on the next tick
				pushBackTime := now.Add(-p.IdleThreshold)

				p.mu.Lock()
				for _, el := range toTryAgain {
					_ = p.pushIdleInOrder(el, pushBackTime)
				}
				p.mu.Unlock()
			}

			var (
				sleep bool
				delay time.Duration
			)
			p.mu.Lock()

			if s, touched := p.peekFirstIdle(); s == nil {
				// No sessions to check. Let the Put() caller to wake up
				// keeper when session arrive.
				sleep = true
				p.keeperWake = wake
			} else {
				// NOTE: negative delay is also fine.
				delay = p.IdleThreshold - now.Sub(touched)
			}

			// Takers notification broadcast channel.
			touchingDone := p.touchingDone
			p.touchingDone = nil
			p.touching = false

			p.mu.Unlock()

			if !sleep {
				timer.Reset(delay)
			}
			for i, s := range toDelete {
				toDelete[i] = nil
				ctx, cancel := context.WithTimeout(
					context.Background(),
					p.DeleteTimeout,
				)
				_ = s.Close(ctx)
				cancel()
			}
			if touchingDone != nil {
				close(touchingDone)
			}
		}
	}
}

// getWaitCh returns pointer to a channel of sessions.
//
// Note that returning a pointer reduces allocations on sync.Pool usage –
// sync.Pool.Get() returns empty interface, which leads to allocation for
// non-pointer values.
func (p *pool) getWaitCh() *chan table.Session {
	if p.testHookGetWaitCh != nil {
		p.testHookGetWaitCh()
	}
	s, ok := p.waitChPool.Get().(*chan table.Session)
	if !ok {
		// NOTE: MUST NOT be buffered.
		// In other case we could cork an already no-owned channel.
		ch := make(chan table.Session)
		s = &ch
	}
	return s
}

// putWaitCh receives pointer to a channel and makes it available for further
// use.
// Note that ch MUST NOT be owned by any goroutine at the call moment and ch
// MUST NOT contain any value.
func (p *pool) putWaitCh(ch *chan table.Session) {
	p.waitChPool.Put(ch)
}

// p.mu must be held.
func (p *pool) peekFirstIdle() (s table.Session, touched time.Time) {
	el := p.idle.Front()
	if el == nil {
		return
	}
	s = el.Value.(table.Session)
	info, has := p.index[s]
	if !has || el != info.idle {
		panicLocked(&p.mu, "ydb: table: inconsistent session pool index")
	}
	return s, info.touched
}

// removes first session from idle and resets the keepAliveCount
// to prevent session from dying in the keeper after it was returned
// to be used only in outgoing functions that make session busy.
// p.mu must be held.
func (p *pool) removeFirstIdle() table.Session {
	s, _ := p.peekFirstIdle()
	if s != nil {
		info := p.removeIdle(s)
		info.keepAliveCount = 0
		p.index[s] = info
	}
	return s
}

// Increments the Keep Alive Counter and returns the previous number.
// Unlike other info modifiers, this one doesn't care if it didn't find the session, it skips
// the action. You can still check it later if needed, if the return code is -1
// p.mu must be held.
func (p *pool) incrementKeepAlive(s table.Session) int {
	info, has := p.index[s]
	if !has {
		return -1
	}
	ret := info.keepAliveCount
	info.keepAliveCount++
	p.index[s] = info
	return ret
}

// p.mu must be held.
func (p *pool) touchCond() <-chan struct{} {
	if p.touchingDone == nil {
		p.touchingDone = make(chan struct{})
	}
	return p.touchingDone
}

// p.mu must be held.
func (p *pool) notify(s table.Session) (notified bool) {
	for el := p.waitq.Front(); el != nil; el = p.waitq.Front() {
		// Some goroutine is waiting for a session.
		//
		// It could be in this states:
		//   1) Reached the select code and awaiting for a value in channel.
		//   2) Reached the select code but already in branch of deadline
		//   cancelation. In this case it is locked on p.mu.Lock().
		//   3) Not reached the select code and thus not reading yet from the
		//   channel.
		//
		// For cases (2) and (3) we close the channel to signal that goroutine
		// missed something and may want to retry (especially for case (3)).
		//
		// After that we taking a next waiter and repeat the same.
		ch := p.waitq.Remove(el).(*chan table.Session)
		select {
		case *ch <- s:
			// Case (1).
			return true
		default:
			// Case (2) or (3).
			close(*ch)
		}
	}
	return false
}

// CloseSession provides the most effective way of session closing
// instead of plain session.Close.
// CloseSession must be fast. If necessary, can be async.
func (p *pool) CloseSession(ctx context.Context, s table.Session) error {
	ctx, cancel := context.WithTimeout(
		deadline.ContextWithoutDeadline(ctx),
		p.DeleteTimeout,
	)
	closeSessionDone := trace.TablePoolOnCloseSession(p.Trace.TablePool, ctx, s.ID())
	go func() {
		defer cancel()
		closeSessionDone(ctx, s.ID(), s.Close(ctx))
	}()
	return nil
}

func (p *pool) keepAliveSession(ctx context.Context, s table.Session) (options.SessionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, p.KeepAliveTimeout)
	defer cancel()
	return s.KeepAlive(ctx)
}

// p.mu must be held.
func (p *pool) removeIdle(s table.Session) sessionInfo {
	info, has := p.index[s]
	if !has || info.idle == nil {
		panicLocked(&p.mu, "ydb: table: inconsistent session pool index")
	}

	p.idle.Remove(info.idle)
	info.idle = nil
	p.index[s] = info
	return info
}

// Removes session from idle pool and resets keepAliveCount for it not
// to die in keeper when it will be returned
// to be used only in outgoing functions that make session busy.
// p.mu must be held.
func (p *pool) takeIdle(s table.Session) (has, took bool) {
	var info sessionInfo
	info, has = p.index[s]
	if !has {
		// Could not be strict here and panic – session may become deleted by
		// keeper().
		return
	}
	if info.idle == nil {
		// session s is not idle.
		return
	}
	took = true
	info = p.removeIdle(s)
	info.keepAliveCount = 0
	p.index[s] = info
	return
}

// p.mu must be held.
func (p *pool) pushIdle(s table.Session, now time.Time) {
	p.handlePushIdle(s, now, p.idle.PushBack(s))
}

// p.mu must be held.
func (p *pool) pushIdleInOrder(s table.Session, now time.Time) (el *list.Element) {
	var prev *list.Element
	for prev = p.idle.Back(); prev != nil; prev = prev.Prev() {
		s := prev.Value.(table.Session)
		t := p.index[s].touched
		if !now.Before(t) { // now >= t
			break
		}
	}
	if prev != nil {
		el = p.idle.InsertAfter(s, prev)
	} else {
		el = p.idle.PushFront(s)
	}
	p.handlePushIdle(s, now, el)
	return el
}

// p.mu must be held.
func (p *pool) pushIdleInOrderAfter(s table.Session, now time.Time, mark *list.Element) *list.Element {
	if mark != nil {
		n := p.idle.Len()
		el := p.idle.InsertAfter(s, mark)
		if n < p.idle.Len() {
			// List changed, thus mark belongs to list.
			p.handlePushIdle(s, now, el)
			return el
		}
	}
	return p.pushIdleInOrder(s, now)
}

// p.mu must be held.
func (p *pool) handlePushIdle(s table.Session, now time.Time, el *list.Element) {
	info, has := p.index[s]
	if !has {
		panicLocked(&p.mu, "ydb: table: trying to store session created outside of the pool")
	}
	if info.idle != nil {
		panicLocked(&p.mu, "ydb: table: inconsistent session pool index")
	}

	info.touched = now
	info.idle = el
	p.index[s] = info

	p.wakeUpKeeper()
}

// p.mu must be held.
func (p *pool) wakeUpKeeper() {
	if wake := p.keeperWake; wake != nil {
		p.keeperWake = nil
		close(wake)
	}
}

type sessionInfo struct {
	idle           *list.Element
	ready          *list.Element
	touched        time.Time
	keepAliveCount int
}

func panicLocked(mu sync.Locker, message string) {
	mu.Unlock()
	panic(message)
}

type poolStats struct {
	Idle             int
	Index            int
	WaitQ            int
	MinSize          int
	MaxSize          int
	CreateInProgress int
}
