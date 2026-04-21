package errx

// 通用错误码 (40001-40099)
const (
	CodeOK             = 0
	CodeInvalidRequest = 40001
)

// 认证授权错误码 (40101-40199)
const (
	CodeUnauthorized        = 40101
	CodeInvalidCredentials = 40102
	CodeTokenExpired       = 40103
	CodeTokenInvalid       = 40104
	CodeForbidden          = 40105
)

// User服务错误码 (40201-40299)
const (
	CodeUserNotFound       = 40201
	CodeUserAlreadyExists  = 40202
	CodeUserCreateFailed   = 40203
	CodeUserUpdateFailed   = 40204
	CodeUserDeleteFailed   = 40205
	CodeUserDisabled       = 40206
	CodeInvalidPassword    = 40207
	CodePasswordMismatch   = 40208
)

// Task服务错误码 (40301-40399)
const (
	CodeTaskNotFound      = 40301
	CodeTaskCreateFailed  = 40302
	CodeTaskUpdateFailed  = 40303
	CodeTaskDeleteFailed  = 40304
	CodeTaskStatusInvalid = 40305
)

// 资源不存在错误码 (40401-40499)
const (
	CodeNotFound           = 40401
	CodeOrderNotFound       = 40402
	CodePaymentNotFound     = 40403
	CodeInventoryNotFound   = 40404
	CodeSkuNotFound         = 40405
	CodeReservationNotFound = 40406
	CodeCouponNotFound      = 40407
	CodeActivityNotFound    = 40408
	CodePriceNotFound       = 40409
	CodeRefundNotFound      = 40410
)

// Order服务错误码 (40501-40599)
const (
	CodeOrderCreateFailed     = 40501
	CodeOrderUpdateFailed     = 40502
	CodeOrderCancelFailed     = 40503
	CodeOrderNotCancelable    = 40504
	CodeInvalidOrderState     = 40505
	CodeInsufficientInventory = 40506
	CodeOrderAmountInvalid    = 40507
	CodeOrderAlreadyPaid      = 40508
	CodeOrderAlreadyCanceled  = 40509
)

// Payment服务错误码 (40601-40699)
const (
	CodePaymentCreateFailed  = 40601
	CodePaymentUpdateFailed  = 40602
	CodePaymentNotPayable   = 40603
	CodeInvalidPaymentState = 40604
	CodePaymentTimeout      = 40605
	CodePaymentCanceled     = 40606
	CodePaymentRefunded     = 40607
	CodePaymentAmountInvalid = 40608
)

// Inventory服务错误码 (40701-40799)
const (
	CodeInventoryReserveFailed = 40701
	CodeInvalidReservationState = 40702
	CodeInventoryRestoreFailed  = 40703
	CodeInventoryInsufficient   = 40704
	CodeInventoryLockFailed     = 40705
	CodeSkuAlreadyExists        = 40706
)

// Activity服务错误码 (40801-40899)
const (
	CodeActivityCreateFailed = 40801
	CodeCouponOutOfStock     = 40802
	CodeCouponAlreadyClaimed = 40803
	CodeCouponExpired        = 40804
	CodeCouponInvalid        = 40805
	CodeSeckillNotStart      = 40806
	CodeSeckillAlreadyEnd     = 40807
	CodeSeckillInsufficient  = 40808
	CodeActivityNotActive    = 40809
)

// Refund服务错误码 (40901-40999)
const (
	CodeRefundCreateFailed   = 40901
	CodeRefundAlreadyProcessed = 40902
	CodeRefundAmountInvalid  = 40903
	CodeRefundExceedAmount   = 40904
	CodeRefundNotAllowed     = 40905
	CodeRefundTimeout        = 40906
)

// Price服务错误码 (41001-41099)
const (
	CodePriceExpired     = 41002
	CodePriceUpdateFailed = 41003
)

// 服务端错误码 (50001-50099)
const (
	CodeInternalError     = 50001
	CodeUpstreamUnavail  = 50201
	CodeDatabaseError    = 50002
	CodeCacheError       = 50003
	CodeRpcCallFailed    = 50004
	CodeSerializeFailed  = 50005
	CodeQueuePublishFailed = 50006
)

// 通用错误消息
const (
	MsgOK               = "OK"
	MsgInvalidRequest   = "invalid request"
	MsgUnauthorized     = "missing authorization"
	MsgForbidden        = "forbidden"
	MsgNotFound         = "not found"
	MsgInternalError    = "internal error"
	MsgUpstreamUnavail  = "upstream unavailable"
)

// 认证授权错误消息
const (
	MsgInvalidCredentials = "invalid username or password"
	MsgTokenExpired       = "token has expired"
	MsgTokenInvalid       = "invalid token"
)

// User服务错误消息
const (
	MsgUserNotFound      = "user not found"
	MsgUserAlreadyExists = "user already exists"
	MsgUserCreateFailed  = "create user failed"
	MsgUserUpdateFailed  = "update user failed"
	MsgUserDeleteFailed  = "delete user failed"
	MsgUserDisabled       = "user account is disabled"
	MsgInvalidPassword   = "invalid password format"
	MsgPasswordMismatch  = "passwords do not match"
)

// Task服务错误消息
const (
	MsgTaskNotFound      = "task not found"
	MsgTaskCreateFailed  = "create task failed"
	MsgTaskUpdateFailed  = "update task failed"
	MsgTaskDeleteFailed  = "delete task failed"
	MsgTaskStatusInvalid = "invalid task status"
)

// Order服务错误消息
const (
	MsgOrderNotFound       = "order not found"
	MsgOrderCreateFailed  = "create order failed"
	MsgOrderUpdateFailed  = "update order failed"
	MsgOrderCancelFailed   = "cancel order failed"
	MsgOrderNotCancelable  = "order cannot be canceled"
	MsgInvalidOrderState   = "invalid order state"
	MsgInsufficientInventory = "insufficient inventory"
	MsgOrderAmountInvalid  = "invalid order amount"
	MsgOrderAlreadyPaid    = "order already paid"
	MsgOrderAlreadyCanceled = "order already canceled"
)

// Payment服务错误消息
const (
	MsgPaymentNotFound      = "payment not found"
	MsgPaymentCreateFailed  = "create payment failed"
	MsgPaymentUpdateFailed  = "update payment failed"
	MsgPaymentNotPayable    = "payment is not payable"
	MsgInvalidPaymentState  = "invalid payment state"
	MsgPaymentTimeout       = "payment timeout"
	MsgPaymentCanceled      = "payment canceled"
	MsgPaymentRefunded      = "payment already refunded"
	MsgPaymentAmountInvalid = "invalid payment amount"
)

// Inventory服务错误消息
const (
	MsgSkuNotFound              = "sku not found"
	MsgInventoryReserveFailed   = "reserve inventory failed"
	MsgInvalidReservationState  = "invalid reservation state"
	MsgInventoryRestoreFailed    = "restore inventory failed"
	MsgInventoryInsufficient     = "insufficient inventory"
	MsgInventoryLockFailed       = "lock inventory failed"
	MsgSkuAlreadyExists          = "sku already exists"
	MsgReservationNotFound       = "reservation not found"
)

// Activity服务错误消息
const (
	MsgCouponNotFound       = "coupon not found"
	MsgCouponOutOfStock     = "coupon out of stock"
	MsgCouponAlreadyClaimed = "coupon already claimed"
	MsgCouponExpired        = "coupon has expired"
	MsgCouponInvalid        = "invalid coupon"
	MsgActivityNotFound     = "activity not found"
	MsgSeckillNotStart      = "seckill has not started"
	MsgSeckillAlreadyEnd    = "seckill has ended"
	MsgSeckillInsufficient  = "seckill inventory insufficient"
	MsgActivityNotActive    = "activity is not active"
	MsgActivityCreateFailed  = "create activity failed"
)

// Refund服务错误消息
const (
	MsgRefundNotFound           = "refund not found"
	MsgRefundCreateFailed       = "create refund failed"
	MsgRefundAlreadyProcessed   = "refund already processed"
	MsgRefundAmountInvalid      = "invalid refund amount"
	MsgRefundExceedAmount       = "refund amount exceeds paid amount"
	MsgRefundNotAllowed         = "refund not allowed"
	MsgRefundTimeout            = "refund timeout"
)

// Price服务错误消息
const (
	MsgPriceNotFound    = "price not found"
	MsgPriceExpired     = "price has expired"
	MsgPriceUpdateFailed = "update price failed"
)

// 服务端错误消息
const (
	MsgDatabaseError      = "database error"
	MsgCacheError         = "cache error"
	MsgRpcCallFailed      = "rpc call failed"
	MsgSerializeFailed    = "serialize failed"
	MsgQueuePublishFailed = "message queue publish failed"
)
