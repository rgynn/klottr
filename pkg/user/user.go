package user

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/rgynn/ptrconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("user not found")

var ErrAlreadyExists = errors.New("user already exists")

type Repository interface {
	Create(ctx context.Context, m *Model) error
	Search(ctx context.Context, username, role *string, from, size int64) ([]*Model, error)
	Get(ctx context.Context, username *string) (*Model, error)
	Delete(ctx context.Context, username, role *string) error

	IncThreadsCounter(ctx context.Context, username *string) error
	DecThreadsCounter(ctx context.Context, username *string) error
	IncCommentsCounter(ctx context.Context, username *string) error
	DecCommentsCounter(ctx context.Context, username *string) error

	IncThreadsVotes(ctx context.Context, username *string) error
	DecThreadsVotes(ctx context.Context, username *string) error
	IncCommentsVotes(ctx context.Context, username *string) error
	DecCommentsVotes(ctx context.Context, username *string) error

	Close() error
}

type Counters struct {
	Num   Counter `json:"num"  bson:"num"`
	Votes Counter `json:"votes"  bson:"votes"`
}

type Counter struct {
	Threads  uint32 `json:"threads"  bson:"threads"`
	Comments uint32 `json:"comments"  bson:"comments"`
}

type Model struct {
	ID           *primitive.ObjectID `json:"id" bson:"_id"`
	Role         *string             `json:"role" bson:"role"`
	Validated    bool                `json:"validated,omitempty"  bson:"validated,omitempty"`
	Username     *string             `json:"username,omitempty"  bson:"username,omitempty"`
	Password     *string             `json:"password,omitempty"  bson:"_"`
	PasswordHash *string             `json:"password_hash,omitempty"  bson:"password_hash,omitempty"`
	Email        *string             `json:"email,omitempty"  bson:"_"`
	EmailHash    *string             `json:"email_hash,omitempty"  bson:"email_hash,omitempty"`
	Counters     Counters            `json:"counters"  bson:"counters"`
	Created      *time.Time          `json:"created"  bson:"created"`
	Updated      *time.Time          `json:"updated,omitempty"  bson:"updated,omitempty"`
}

func (m *Model) HashPassword() error {

	if m == nil {
		return errors.New("no m *Model provided")
	}

	if m.Password == nil {
		return errors.New("no m.Password provided")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*m.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	m.Password = nil
	m.PasswordHash = ptrconv.StringPtr(string(hashedPassword))

	return nil
}

func (m *Model) HashEmail() error {

	if m == nil {
		return errors.New("no m *Model provided")
	}

	if m.Email == nil {
		return nil
	}

	hashedEmail, err := bcrypt.GenerateFromPassword([]byte(*m.Email), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	m.Email = nil
	m.EmailHash = ptrconv.StringPtr(string(hashedEmail))

	return nil
}

func (m *Model) ValidForSave() error {

	if m == nil {
		return errors.New("no m *user.Model provided")
	}

	if m.ID != nil {
		return errors.New("cannot provide m.ID for new user")
	}

	if m.Role == nil {
		return errors.New("no m.Role provided")
	}

	switch *m.Role {
	case "user", "admin":
		break
	default:
		return errors.New("invalid m.Role provided")
	}

	if m.Username == nil {
		return errors.New("no m.Username provided")
	}

	if m.Username != nil && utf8.RuneCountInString(*m.Username) > 255 {
		return errors.New("username too long")
	}

	if m.Password != nil {
		return errors.New("cannot provide raw m.Password when saving, make sure to hash it first")
	}

	if m.PasswordHash == nil {
		return errors.New("no m.PasswordHash provided")
	}

	if m.Email != nil {
		return errors.New("cannot provide raw m.Email when saving, make sure to hash it first")
	}

	if m.Created != nil {
		return errors.New("cannot prvide m.Created for new user")
	}

	if m.Updated != nil {
		return errors.New("cannot prvide m.Updated for new user")
	}

	return nil
}

func (m *Model) ValidPassword(password *string) error {

	if m == nil {
		return errors.New("no m *user.Model provided")
	}

	if password == nil {
		return errors.New("no password provided")
	}

	if m.PasswordHash == nil {
		return errors.New("no m.PasswordHash provided")
	}

	return bcrypt.CompareHashAndPassword([]byte(*m.PasswordHash), []byte(*password))
}

func (m *Model) ValidEmail(email *string) error {

	if m == nil {
		return errors.New("no m *user.Model provided")
	}

	if email == nil {
		return errors.New("no email provided")
	}

	if m.EmailHash == nil {
		return errors.New("no m.EmailHash provided")
	}

	return bcrypt.CompareHashAndPassword([]byte(*m.EmailHash), []byte(*email))
}
