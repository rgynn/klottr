package thread

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rgynn/klottr/pkg/helper"
	"github.com/rgynn/ptrconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrCategoryNotFound = errors.New("thread category not found")

var ErrNotFound = errors.New("thread not found")

type Repository interface {
	Create(ctx context.Context, m *Model) error
	Get(ctx context.Context, id *string) (*Model, error)
	List(ctx context.Context, from, size int64) ([]*Model, error)
	Delete(ctx context.Context, id *string) error

	IncVote(ctx context.Context, id *string) error
	DecVote(ctx context.Context, id *string) error
	IncComments(ctx context.Context, id *string) error
	DecComments(ctx context.Context, id *string) error

	Close() error
}

type Counters struct {
	Votes    int64  `json:"votes"  bson:"votes"`
	Comments uint32 `json:"comments"  bson:"comments"`
}

type Model struct {
	ID        *primitive.ObjectID `json:"id" bson:"_id"`
	UserID    *primitive.ObjectID `json:"user_id"  bson:"user_id,omitempty"`
	SlugID    *string             `json:"slug_id"  bson:"slug_id"`
	SlugTitle *string             `json:"slug_title"  bson:"slug_title"`
	Category  *string             `json:"category"  bson:"category"`
	Title     *string             `json:"title,omitempty"  bson:"title,omitempty"`
	URL       *string             `json:"url,omitempty"  bson:"url,omitempty"`
	Content   string              `json:"content"  bson:"content"`
	Counters  Counters            `json:"counters"  bson:"counters"`
	Created   *time.Time          `json:"created"  bson:"created"`
	Updated   *time.Time          `json:"updated"  bson:"updated"`
}

func (m *Model) ValidForSave() error {

	if m == nil {
		return errors.New("no m *thread.Model provided")
	}

	if m.ID != nil {
		return errors.New("cannot provide new object id for thread")
	}

	if m.UserID == nil {
		return errors.New("no m.UserID provided for new thread")
	}

	if m.Category == nil {
		return errors.New("no m.Category provided for new thread")
	}

	if m.Title == nil {
		return errors.New("no m.Title provided")
	}

	if m.SlugID == nil {
		return errors.New("no m.SlugID provided")
	}

	if m.SlugTitle == nil {
		return errors.New("no m.SlugTitle provided")
	}

	if m.Title != nil {
		if utf8.RuneCountInString(ptrconv.StringPtrString(m.Title)) > 300 {
			return errors.New("m.Title too long")
		}
	}

	if m.Content == "" {
		return errors.New("no m.Content provided")
	}

	if m.Content != "" {
		if utf8.RuneCountInString(m.Content) > 3000 {
			return errors.New("content too long")
		}
	}

	if m.Created == nil || m.Created.IsZero() {
		return errors.New("no m.Posted provided")
	}

	return nil
}

func (m *Model) GenerateSlugs() error {

	if m == nil {
		return errors.New("no m *thread.Model provided")
	}

	if m.Title == nil {
		return errors.New("no m.Title provided")
	}

	m.SlugID = ptrconv.StringPtr(helper.RandomString(5))
	m.SlugTitle = ptrconv.StringPtr(fmt.Sprintf("%.47s", strings.ReplaceAll(*m.Title, " ", "_")))

	return nil
}
