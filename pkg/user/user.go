package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/rgynn/ptrconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var ErrDeactivated = errors.New("user account deactivated")

var ErrNotFound = errors.New("user not found")

var ErrAlreadyExists = errors.New("user already exists")

type Repository interface {
	Create(ctx context.Context, m *Model) error
	Search(ctx context.Context, username, role *string, from, size int64) ([]*Model, error)
	GetByID(ctx context.Context, id *string) (*Model, error)
	GetByUsername(ctx context.Context, username *string) (*Model, error)
	Deactivate(ctx context.Context, username, role *string) error
	Delete(ctx context.Context, username, role *string) error
	IncCounter(ctx context.Context, username, field *string, value int8) error
	UpsertVote(ctx context.Context, username *string, vote *Vote) error
}

type Counters struct {
	Num   Counter `json:"num"  bson:"num"`
	Votes Counter `json:"votes"  bson:"votes"`
}

type Counter struct {
	Threads  uint32 `json:"threads"  bson:"threads"`
	Comments uint32 `json:"comments"  bson:"comments"`
}

type Votes struct {
	Threads  map[string]int8 `json:"threads" bson:"threads"`
	Comments map[string]int8 `json:"comments" bson:"comments"`
}

type Vote struct {
	SlugType *string `json:"slug_type" bson:"slug_type"`
	SlugID   *string `json:"slug_id" bson:"slug_id"`
	Value    *int8   `json:"value" bson:"value"`
}

func (v *Vote) ValidForSave() error {

	if v == nil {
		return errors.New("no vote provided")
	}

	if v.SlugType == nil {
		return errors.New("no v.SlugType provided")
	}

	switch *v.SlugType {
	case "comments", "threads":
		break
	default:
		return fmt.Errorf("invalid v.SlugType provided: %s", *v.SlugType)
	}

	if v.SlugID == nil {
		return errors.New("no v.SlugID provided")
	}

	if v.Value == nil {
		return errors.New("no v.Value provided")
	}

	if *v.Value < -1 || *v.Value > 1 {
		return errors.New("v.Value cannot be lower than -1 or greater than 1")
	}

	return nil
}

type Model struct {
	ID           *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Role         *string             `json:"role" bson:"role"`
	Validated    bool                `json:"validated"  bson:"validated"`
	Username     *string             `json:"username"  bson:"username"`
	Password     *string             `json:"password,omitempty"  bson:"password,omitempty"`
	PasswordHash *string             `json:"password_hash,omitempty"  bson:"password_hash,omitempty"`
	Email        *string             `json:"email,omitempty"  bson:"email,omitempty"`
	EmailHash    *string             `json:"email_hash,omitempty"  bson:"email_hash,omitempty"`
	Counters     Counters            `json:"counters"  bson:"counters"`
	Votes        Votes               `json:"votes" bson:"votes"`
	Created      *time.Time          `json:"created"  bson:"created"`
	Updated      *time.Time          `json:"updated,omitempty"  bson:"updated,omitempty"`
	Deactivated  *time.Time          `json:"deactivated,omitempty"  bson:"deactivated,omitempty"`
}

func (m *Model) IsDeactivated() bool {
	return m.Deactivated != nil
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
