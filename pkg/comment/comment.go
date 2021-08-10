package comment

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrNotFound = errors.New("comment not found")

type Repository interface {
	Create(ctx context.Context, m *Model) error
	Get(ctx context.Context, id *string) (*Model, error)
	ListByThreadID(ctx context.Context, threadID *string, from, size int64) ([]*Model, error)
	ListByUserID(ctx context.Context, userID *string, from, size int64) ([]*Model, error)
	Delete(ctx context.Context, id *string) error

	IncVotes(ctx context.Context, id *string) error
	DecVotes(ctx context.Context, id *string) error

	Close() error
}

type Model struct {
	ID        *primitive.ObjectID `json:"id" bson:"_id"`
	ThreadID  *primitive.ObjectID `json:"thread_id" bson:"thread_id"`
	UserID    *primitive.ObjectID `json:"user_id"  bson:"user_id"`
	ReplyToID *primitive.ObjectID `json:"reply_to_id,omitempty"  bson:"reply_to_id,omitempty"`
	Content   string              `json:"content"  bson:"content"`
	Posted    time.Time           `json:"posted"  bson:"posted"`
	Votes     int64               `json:"votes"  bson:"votes"`
}

func (m *Model) ValidForSave() error {

	if m == nil {
		return errors.New("no m *thread.Model provided")
	}

	if m.ID != nil {
		return errors.New("cannot provide m.ID for new thread")
	}

	if m.ThreadID == nil {
		return errors.New("no m.ThreadID provided for new thread")
	}

	if m.UserID == nil {
		return errors.New("no m.UserID provided for new thread")
	}

	if m.Content == "" {
		return errors.New("no m.Content provided")
	}

	if m.Content != "" {
		if utf8.RuneCountInString(m.Content) > 3000 {
			return errors.New("comment too long")
		}
	}

	if m.Posted.IsZero() {
		return errors.New("no m.Posted provided")
	}

	return nil
}
