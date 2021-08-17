package comment

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/rgynn/klottr/pkg/helper"
	"github.com/rgynn/ptrconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrNotFound = errors.New("comment not found")

type Repository interface {
	Create(ctx context.Context, m *Model) error
	Get(ctx context.Context, slugID *string) (*Model, error)
	ListByThreadID(ctx context.Context, threadID *primitive.ObjectID, from, size int64) ([]*Model, error)
	ListByUsername(ctx context.Context, username *string, from, size int64) ([]*Model, error)
	Delete(ctx context.Context, slugID *string) error

	IncVotes(ctx context.Context, slugID *string, value int8) error
}

type Model struct {
	ID        *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ThreadID  *primitive.ObjectID `json:"thread_id,omitempty" bson:"thread_id,omitempty"`
	ReplyToID *primitive.ObjectID `json:"reply_to_id,omitempty"  bson:"reply_to_id,omitempty"`
	SlugID    *string             `json:"slug_id,omitempty"  bson:"slug_id,omitempty"`
	Username  *string             `json:"username,omitempty"  bson:"username,omitempty"`
	Content   string              `json:"content"  bson:"content"`
	Votes     int64               `json:"votes"  bson:"votes"`
	Updated   *time.Time          `json:"updated,omitempty"  bson:"updated,omitempty"`
	Created   time.Time           `json:"created"  bson:"created"`
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

	if m.Username == nil {
		return errors.New("no m.Username provided for new thread")
	}

	if m.Content == "" {
		return errors.New("no m.Content provided")
	}

	if m.Content != "" {
		if utf8.RuneCountInString(m.Content) > 3000 {
			return errors.New("comment too long")
		}
	}

	if m.Votes != 0 {
		return errors.New("cannot provide num votes when creating new comment")
	}

	if m.Created.IsZero() {
		return errors.New("no m.Posted provided")
	}

	return nil
}

func (m *Model) GenerateSlugs() error {

	if m == nil {
		return errors.New("no m *thread.Model provided")
	}

	m.SlugID = ptrconv.StringPtr(helper.RandomString(10))

	return nil
}
