package v1

import (
	"context"
	"regexp"
	"time"

	"github.com/go-faster/errors"
	"github.com/kaynelza/NTF/internal/infrastructure/helper"
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
	CreateNewNotification(ctx context.Context, notification helper.Notification) (string, time.Time, error)
	CancelNotification(ctx context.Context, id string) (helper.Status, error)
	GetInfoAboutNotification(ctx context.Context, id string) (helper.Notification, error)
	ListAllNotifications(ctx context.Context, email string, status helper.Status, limit, offset int) ([]helper.Notification, int, error)
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
		return nil, errors.Wrap(err, "in create")
	}

	id, sendAt, err := n.repo.CreateNewNotification(ctx, helper.Notification{
		Recipient: request.Recipient,
		Title:     request.Title,
		Body:      request.Body,
		SendAt:    request.GetSendAt().AsTime(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create new notification")
	}

	return &v1.CreateResponse{
		Id:     id,
		SendAt: timestamppb.New(sendAt),
	}, nil
}

func (n *NotificationServiceServer) Cancel(ctx context.Context, request *v1.CancelRequest) (*v1.CancelResponse, error) {
	if err := n.validateId(request.Id); err != nil {
		return nil, errors.Wrap(err, "in cancel")
	}

	state, err := n.repo.CancelNotification(ctx, request.Id)
	if err != nil {
		return nil, errors.Wrap(err, "cancel notification")
	}

	return &v1.CancelResponse{Status: helper.StatusToPB(state)}, nil
}

func (n *NotificationServiceServer) Get(ctx context.Context, request *v1.GetRequest) (*v1.Notification, error) {
	if err := n.validateId(request.Id); err != nil {
		return nil, errors.Wrap(err, "in get")
	}

	notification, err := n.repo.GetInfoAboutNotification(ctx, request.Id)
	if err != nil {
		return nil, errors.Wrap(err, "get info about notification")
	}

	return helper.NotificationToPB(notification), nil
}

func (n *NotificationServiceServer) List(ctx context.Context, request *v1.ListRequest) (*v1.ListResponse, error) {
	if !n.emailRegexp.MatchString(request.Recipient) {
		return nil, errors.New("email is not valid")
	}

	list, total, err := n.repo.ListAllNotifications(ctx, request.Recipient, helper.StatusFromPB(request.Status), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, errors.Wrap(err, "list all notifications")
	}

	var listPB []*v1.Notification
	for _, v := range list {
		listPB = append(listPB, helper.NotificationToPB(v))
	}

	return &v1.ListResponse{
		Items: listPB,
		Total: int32(total),
	}, nil
}

func (n *NotificationServiceServer) validateCreateReq(req *v1.CreateRequest) error {
	if req.GetDelay() != nil {
		if req.GetDelay().AsDuration() > time.Millisecond {
			return status.Error(codes.InvalidArgument, "delay is too short")
		}
	}

	if req.GetSendAt() != nil {
		if req.GetSendAt().AsTime().After(time.Now().Add(time.Millisecond)) {
			return status.Error(codes.InvalidArgument, "send at time is passed")
		}
	}

	if !n.emailRegexp.MatchString(req.GetRecipient()) {
		return status.Error(codes.InvalidArgument, "email is invalid")
	}

	if req.GetTitle() == "" {
		return status.Error(codes.InvalidArgument, "title is empty")
	}

	return nil

}

func (n *NotificationServiceServer) validateId(id string) error {
	if id == "" {
		return errors.New("id is invalid")
	}
	return nil
}
