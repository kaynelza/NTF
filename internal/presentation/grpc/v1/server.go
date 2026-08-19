package v1

import (
	"context"
	"regexp"
	"time"

	"github.com/go-faster/errors"
	"github.com/kaynelza/NTF/internal/infrastructure/entity"
	v1 "github.com/kaynelza/NTF/pkg/grpc/notifyd/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationServiceServer struct {
	v1.UnimplementedNotificationServiceServer
	emailRegexp *regexp.Regexp
	repo        Storage
}

type Storage interface {
	CreateNotification(ctx context.Context, notification entity.Notification) (id string, sendAt time.Time, err error)
	GetNotificationByID(ctx context.Context, id string) (notification entity.Notification, err error)
	ListAllNotifications(ctx context.Context, recipient string, status entity.Status, limit, offset int) (list []entity.Notification, total int, err error)
	CancelNotification(ctx context.Context, id string) error
}

func New() (*NotificationServiceServer, error) {
	eReg, err := regexp.Compile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")
	if err != nil {
		return nil, errors.Wrap(err, "regexp email")
	}

	return &NotificationServiceServer{
		emailRegexp: eReg,
	}, nil
}

func (n *NotificationServiceServer) Create(ctx context.Context, request *v1.CreateRequest) (*v1.CreateResponse, error) {
	if err := n.validateCreateReq(request); err != nil {
		return nil, n.newError(errors.Wrap(err, "request validation"))
	}

	id, sendAt, err := n.repo.CreateNotification(ctx, entity.Notification{
		Recipient: request.Recipient,
		Title:     request.Title,
		Body:      request.Body,
		SendAt:    request.GetSendAt().AsTime(),
	})
	if err != nil {
		return nil, n.newError(errors.Wrap(err, "create new notification"))
	}

	return &v1.CreateResponse{
		Id:     id,
		SendAt: timestamppb.New(sendAt),
	}, nil
}

func (n *NotificationServiceServer) Cancel(ctx context.Context, request *v1.CancelRequest) (*v1.CancelResponse, error) {
	if err := n.validateCancelReq(request); err != nil {
		return nil, n.newError(errors.Wrap(err, "request validation"))
	}

	notification, err := n.repo.GetNotificationByID(ctx, request.Id)
	if err != nil {
		return nil, n.newError(errors.Wrap(err, "cancel notification"))
	}

	if !entity.CanCancel(notification) {
		return nil, n.newError(entity.ErrFailedPrecondition)
	}

	if err := n.repo.CancelNotification(ctx, notification.Id); err != nil {
		return nil, n.newError(errors.Wrap(err, "failed to cancel notification"))
	}

	return &v1.CancelResponse{Status: entity.StatusToPB(notification.Status)}, nil
}

func (n *NotificationServiceServer) Get(ctx context.Context, request *v1.GetRequest) (*v1.Notification, error) {
	if err := n.validateGetReq(request); err != nil {
		return nil, n.newError(errors.Wrap(err, "request validation"))
	}

	notification, err := n.repo.GetNotificationByID(ctx, request.Id)
	if err != nil {
		return nil, n.newError(errors.Wrap(err, "get info about notification"))
	}

	return entity.NotificationToPB(&notification), nil
}

func (n *NotificationServiceServer) List(ctx context.Context, request *v1.ListRequest) (*v1.ListResponse, error) {
	err := n.validateListReq(request)
	if err != nil {
		return nil, n.newError(errors.Wrap(err, "request validation"))
	}

	list, total, err := n.repo.ListAllNotifications(ctx, request.Recipient, entity.StatusFromPB(request.Status), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, n.newError(errors.Wrap(err, "list all notifications"))
	}

	var listPB []*v1.Notification
	for _, v := range list {
		listPB = append(listPB, entity.NotificationToPB(&v))
	}

	return &v1.ListResponse{
		Items: listPB,
		Total: int32(total),
	}, nil
}

func (n *NotificationServiceServer) validateCreateReq(req *v1.CreateRequest) error {
	if req.GetDelay() != nil {
		if req.GetDelay().AsDuration() > time.Millisecond {
			return errors.Wrap(entity.ErrInvalidArgument, "delay is too short")
		}
	}

	if req.GetSendAt() != nil {
		if req.GetSendAt().AsTime().After(time.Now().Add(time.Millisecond)) {
			return errors.Wrap(entity.ErrInvalidArgument, "send at time is passed")
		}
	}

	if !n.emailRegexp.MatchString(req.GetRecipient()) {
		return errors.Wrap(entity.ErrInvalidArgument, "email is invalid")
	}

	if req.GetTitle() == "" {
		return errors.Wrap(entity.ErrInvalidArgument, "title is empty")
	}

	return nil

}

func (n *NotificationServiceServer) validateCancelReq(request *v1.CancelRequest) error {
	if request.Id == "" {
		return errors.Wrap(entity.ErrInvalidArgument, "id is invalid")
	}
	return nil
}

func (n *NotificationServiceServer) validateGetReq(request *v1.GetRequest) error {
	if request.Id == "" {
		return errors.Wrap(entity.ErrInvalidArgument, "id is invalid")
	}
	return nil
}

func (n *NotificationServiceServer) validateListReq(request *v1.ListRequest) error {
	if !n.emailRegexp.MatchString(request.Recipient) {
		return errors.Wrap(entity.ErrInvalidArgument, "email is invalid")
	}

	if request.Limit > 1000 {
		return errors.Wrap(entity.ErrInvalidArgument, "given limit is over 1000")
	}
	return nil
}

func (n *NotificationServiceServer) newError(err error) error {
	switch {
	case errors.Is(err, entity.ErrAborted):
		return status.Error(codes.Aborted, err.Error())

	case errors.Is(err, entity.ErrCanceled):
		return status.Error(codes.Canceled, err.Error())

	case errors.Is(err, entity.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, entity.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, entity.ErrDataLoss):
		return status.Error(codes.DataLoss, err.Error())

	case errors.Is(err, entity.ErrDeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())

	case errors.Is(err, entity.ErrFailedPrecondition):
		return status.Error(codes.FailedPrecondition, err.Error())

	case errors.Is(err, entity.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, entity.ErrInternal):
		return status.Error(codes.Internal, err.Error())

	case errors.Is(err, entity.ErrOutOfRange):
		return status.Error(codes.OutOfRange, err.Error())

	case errors.Is(err, entity.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, entity.ErrResourceExhausted):
		return status.Error(codes.ResourceExhausted, err.Error())

	case errors.Is(err, entity.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, entity.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())

	case errors.Is(err, entity.ErrUnimplemented):
		return status.Error(codes.Unimplemented, err.Error())

	default:
		return status.Error(codes.Unknown, err.Error())

	}
}
