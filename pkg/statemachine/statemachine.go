package statemachine

import (
	"fmt"
	"time"
)

type State string

type Event string

type Transition struct {
	From  State
	To    State
	Event Event
}

type StateMachine struct {
	name        string
	transitions map[State]map[Event]State
}

func New(name string) *StateMachine {
	return &StateMachine{
		name:        name,
		transitions: make(map[State]map[Event]State),
	}
}

func (sm *StateMachine) AddTransition(from, to State, events ...Event) {
	if sm.transitions[from] == nil {
		sm.transitions[from] = make(map[Event]State)
	}
	for _, e := range events {
		sm.transitions[from][e] = to
	}
}

func (sm *StateMachine) CanTransition(from State, event Event) (bool, State) {
	if transitions, ok := sm.transitions[from]; ok {
		if to, ok := transitions[event]; ok {
			return true, to
		}
	}
	return false, ""
}

func (sm *StateMachine) Transition(from State, event Event) (State, error) {
	valid, to := sm.CanTransition(from, event)
	if !valid {
		return from, fmt.Errorf("%s: invalid transition from %s with event %s", sm.name, from, event)
	}
	return to, nil
}

func (sm *StateMachine) ValidTransitions(from State) map[Event]State {
	return sm.transitions[from]
}

const (
	OrderPending   State = "PENDING"
	OrderReserved  State = "RESERVED"
	OrderPaid     State = "PAID"
	OrderShipped  State = "SHIPPED"
	OrderCompleted State = "COMPLETED"
	OrderCanceled State = "CANCELED"
	OrderExpired  State = "EXPIRED"
	OrderFailed   State = "FAILED"
)

const (
	OrderEventCreate      Event = "CREATE"
	OrderEventReserve     Event = "RESERVE"
	OrderEventPay         Event = "PAY"
	OrderEventShip        Event = "SHIP"
	OrderEventComplete    Event = "COMPLETE"
	OrderEventCancel      Event = "CANCEL"
	OrderEventExpire      Event = "EXPIRE"
	OrderEventFail        Event = "FAIL"
	OrderEventTimeout     Event = "TIMEOUT"
	OrderEventCompensate  Event = "COMPENSATE"
)

func NewOrderStateMachine() *StateMachine {
	sm := New("Order")
	sm.AddTransition(OrderPending, OrderReserved, OrderEventReserve)
	sm.AddTransition(OrderPending, OrderCanceled, OrderEventCancel, OrderEventExpire, OrderEventTimeout)
	sm.AddTransition(OrderPending, OrderFailed, OrderEventFail)

	sm.AddTransition(OrderReserved, OrderPaid, OrderEventPay)
	sm.AddTransition(OrderReserved, OrderCanceled, OrderEventCancel, OrderEventCompensate)
	sm.AddTransition(OrderReserved, OrderExpired, OrderEventExpire, OrderEventTimeout)

	sm.AddTransition(OrderPaid, OrderShipped, OrderEventShip)
	sm.AddTransition(OrderPaid, OrderCanceled, OrderEventCancel, OrderEventCompensate)
	sm.AddTransition(OrderPaid, OrderExpired, OrderEventExpire)

	sm.AddTransition(OrderShipped, OrderCompleted, OrderEventComplete)

	return sm
}

const (
	PaymentPending   State = "PENDING"
	PaymentSuccess  State = "SUCCESS"
	PaymentFailed   State = "FAILED"
	PaymentTimeout  State = "TIMEOUT"
	PaymentCanceled State = "CANCELED"
	PaymentRefunded State = "REFUNDED"
)

const (
	PaymentEventCreate   Event = "CREATE"
	PaymentEventSuccess   Event = "SUCCESS"
	PaymentEventFail      Event = "FAIL"
	PaymentEventTimeout   Event = "TIMEOUT"
	PaymentEventCancel    Event = "CANCEL"
	PaymentEventRefund    Event = "REFUND"
)

func NewPaymentStateMachine() *StateMachine {
	sm := New("Payment")
	sm.AddTransition(PaymentPending, PaymentSuccess, PaymentEventSuccess)
	sm.AddTransition(PaymentPending, PaymentFailed, PaymentEventFail)
	sm.AddTransition(PaymentPending, PaymentTimeout, PaymentEventTimeout)
	sm.AddTransition(PaymentPending, PaymentCanceled, PaymentEventCancel)

	sm.AddTransition(PaymentSuccess, PaymentRefunded, PaymentEventRefund)
	sm.AddTransition(PaymentSuccess, PaymentCanceled, PaymentEventCancel)

	return sm
}

const (
	InventoryAvailable State = "AVAILABLE"
	InventoryReserved  State = "RESERVED"
	InventoryLocked    State = "LOCKED"
	InventoryDepleted  State = "DEPLETED"
)

const (
	InventoryEventReserve   Event = "RESERVE"
	InventoryEventConfirm   Event = "CONFIRM"
	InventoryEventRelease   Event = "RELEASE"
	InventoryEventLock      Event = "LOCK"
	InventoryEventUnlock    Event = "UNLOCK"
	InventoryEventDeplete   Event = "DEPLETE"
	InventoryEventRestore   Event = "RESTORE"
	InventoryEventCancel    Event = "CANCEL"
)

func NewInventoryStateMachine() *StateMachine {
	sm := New("Inventory")
	sm.AddTransition(InventoryAvailable, InventoryReserved, InventoryEventReserve)
	sm.AddTransition(InventoryAvailable, InventoryDepleted, InventoryEventDeplete)
	sm.AddTransition(InventoryAvailable, InventoryLocked, InventoryEventLock)

	sm.AddTransition(InventoryReserved, InventoryAvailable, InventoryEventRelease, InventoryEventCancel)
	sm.AddTransition(InventoryReserved, InventoryAvailable, InventoryEventConfirm)
	sm.AddTransition(InventoryReserved, InventoryDepleted, InventoryEventDeplete)

	sm.AddTransition(InventoryLocked, InventoryAvailable, InventoryEventUnlock)
	sm.AddTransition(InventoryLocked, InventoryDepleted, InventoryEventDeplete)

	sm.AddTransition(InventoryDepleted, InventoryAvailable, InventoryEventRestore)

	return sm
}

const (
	ReservationPending   State = "PENDING"
	ReservationReserved State = "RESERVED"
	ReservationConfirmed State = "CONFIRMED"
	ReservationReleased  State = "RELEASED"
	ReservationExpired   State = "EXPIRED"
)

const (
	ReservationEventReserve   Event = "RESERVE"
	ReservationEventConfirm   Event = "CONFIRM"
	ReservationEventRelease  Event = "RELEASE"
	ReservationEventExpire   Event = "EXPIRE"
	ReservationEventCompensate Event = "COMPENSATE"
)

func NewReservationStateMachine() *StateMachine {
	sm := New("Reservation")
	sm.AddTransition(ReservationPending, ReservationReserved, ReservationEventReserve)
	sm.AddTransition(ReservationPending, ReservationExpired, ReservationEventExpire)

	sm.AddTransition(ReservationReserved, ReservationConfirmed, ReservationEventConfirm)
	sm.AddTransition(ReservationReserved, ReservationReleased, ReservationEventRelease, ReservationEventCompensate)
	sm.AddTransition(ReservationReserved, ReservationExpired, ReservationEventExpire)

	return sm
}

const (
	RefundPending    State = "PENDING"
	RefundProcessing State = "PROCESSING"
	RefundSuccess    State = "SUCCESS"
	RefundFailed     State = "FAILED"
)

const (
	RefundEventInitiate   Event = "INITIATE"
	RefundEventProcess     Event = "PROCESS"
	RefundEventSuccess     Event = "SUCCESS"
	RefundEventFail        Event = "FAIL"
	RefundEventRetry       Event = "RETRY"
)

func NewRefundStateMachine() *StateMachine {
	sm := New("Refund")
	sm.AddTransition(RefundPending, RefundProcessing, RefundEventProcess)
	sm.AddTransition(RefundPending, RefundFailed, RefundEventFail)

	sm.AddTransition(RefundProcessing, RefundSuccess, RefundEventSuccess)
	sm.AddTransition(RefundProcessing, RefundFailed, RefundEventFail)
	sm.AddTransition(RefundProcessing, RefundPending, RefundEventRetry)

	return sm
}

const (
	TaskPending   State = "PENDING"
	TaskRunning   State = "RUNNING"
	TaskSuccess   State = "SUCCESS"
	TaskFailed    State = "FAILED"
	TaskRetry     State = "RETRY"
	TaskDead      State = "DEAD"
)

const (
	TaskEventSubmit   Event = "SUBMIT"
	TaskEventStart    Event = "START"
	TaskEventSuccess  Event = "SUCCESS"
	TaskEventFail     Event = "FAIL"
	TaskEventRetry    Event = "RETRY"
	TaskEventDead     Event = "DEAD"
)

func NewTaskStateMachine() *StateMachine {
	sm := New("Task")
	sm.AddTransition(TaskPending, TaskRunning, TaskEventStart)
	sm.AddTransition(TaskPending, TaskFailed, TaskEventFail)

	sm.AddTransition(TaskRunning, TaskSuccess, TaskEventSuccess)
	sm.AddTransition(TaskRunning, TaskFailed, TaskEventFail)
	sm.AddTransition(TaskRunning, TaskRetry, TaskEventRetry)

	sm.AddTransition(TaskRetry, TaskRunning, TaskEventStart)
	sm.AddTransition(TaskRetry, TaskDead, TaskEventDead)

	sm.AddTransition(TaskFailed, TaskRetry, TaskEventRetry)
	sm.AddTransition(TaskFailed, TaskDead, TaskEventDead)

	return sm
}

type AuditLog struct {
	Entity     string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	FromState State     `json:"from_state"`
	ToState   State     `json:"to_state"`
	Event     Event     `json:"event"`
	Operator  string    `json:"operator"`
	Reason    string    `json:"reason"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (a *AuditLog) Validate() error {
	if a.Entity == "" {
		return fmt.Errorf("entity is required")
	}
	if a.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	return nil
}

func ValidateStateTransition(sm *StateMachine, from State, event Event) error {
	valid, _ := sm.CanTransition(from, event)
	if !valid {
		return fmt.Errorf("invalid transition: from %s with event %s", from, event)
	}
	return nil
}
