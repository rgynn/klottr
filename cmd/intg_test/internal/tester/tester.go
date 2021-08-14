package tester

import (
	"net/http"
	"time"

	"github.com/rgynn/klottr/pkg/config"
	"github.com/sirupsen/logrus"
)

type Tester struct {
	cfg        *config.Config
	logger     *logrus.Logger
	client     *http.Client
	username   string
	password   string
	categories []string
}

func NewTester(cfg *config.Config, logger *logrus.Logger) (*Tester, error) {
	return &Tester{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		username:   "testuser",
		password:   "testpsswd",
		categories: []string{"misc"},
	}, nil
}

func (tester *Tester) Run() error {

	tester.logger.Infof("Integration test started")

	// Test signup and signin

	if err := tester.signupTestUser(); err != nil {
		return err
	}

	token, err := tester.signinTestUser(http.StatusOK)
	if err != nil {
		return err
	}

	if err := tester.validateJWT(token); err != nil {
		return err
	}

	// Test threads

	for _, category := range tester.categories {
		thrd, err := tester.createThread(token, category)
		if err != nil {
			return err
		}
		if err := tester.listThreads(token, category); err != nil {
			return err
		}
		if err := tester.getThread(token, category, thrd.SlugID, thrd.SlugTitle); err != nil {
			return err
		}
		if err := tester.upvoteThread(token, category, thrd.SlugID, thrd.SlugTitle); err != nil {
			return err
		}
		if err := tester.downvoteThread(token, category, thrd.SlugID, thrd.SlugTitle); err != nil {
			return err
		}
		if err := tester.validateVotes(token, category, thrd.SlugID, thrd.SlugTitle); err != nil {
			return err
		}
		cmnt, err := tester.createComment(token, category, thrd.SlugID, thrd.SlugTitle)
		if err != nil {
			return err
		}
		if err := tester.upvoteComment(token, category, thrd.SlugID, thrd.SlugTitle, cmnt.SlugID); err != nil {
			return err
		}
		if err := tester.downvoteComment(token, category, thrd.SlugID, thrd.SlugTitle, cmnt.SlugID); err != nil {
			return err
		}
		if err := tester.validateCommentVotes(token, category, thrd.SlugID, thrd.SlugTitle, cmnt.SlugID); err != nil {
			return err
		}
	}

	// Test deactivate user

	if err := tester.deactivateTestUser(token); err != nil {
		return err
	}

	if _, err = tester.signinTestUser(http.StatusUnauthorized); err != nil {
		return err
	}

	return nil
}
