package tester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/rgynn/klottr/pkg/api"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (tester *Tester) signupTestUser() error {

	url := fmt.Sprintf("http://%s/api/1.0/auth/signup", tester.cfg.Addr)

	reqbody, err := json.Marshal(&user.Model{
		Username: ptrconv.StringPtr(tester.username),
		Password: ptrconv.StringPtr(tester.password),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return err
	}

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		tester.logger.Infof("OK: Signed up username: %s\n", tester.username)
		return nil
	case http.StatusConflict:
		tester.logger.Warnf("WARN: Conflict, user already present: %s\n", tester.username)
		return nil
	default:
		return fmt.Errorf("expected status %d in signup response, got: %d, response body: %s", http.StatusCreated, resp.StatusCode, string(body))
	}
}

func (tester *Tester) signinTestUser(expectedStatusCode int) (*string, error) {

	url := fmt.Sprintf("http://%s/api/1.0/auth/signin", tester.cfg.Addr)

	reqbody, err := json.Marshal(&api.LoginInput{
		Username: ptrconv.StringPtr(tester.username),
		Password: ptrconv.StringPtr(tester.password),
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqbody))
	if err != nil {
		return nil, err
	}

	resp, err := tester.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("expected status %d in signin response, got: %d, response body: %s", http.StatusOK, resp.StatusCode, string(body))
	}

	if expectedStatusCode == http.StatusUnauthorized {
		tester.logger.Infof("OK: Could not sign in deactivated user: %s", tester.username)
		return nil, nil
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	token, ok := response["token"].(string)
	if !ok {
		return nil, fmt.Errorf("expected response to contain a token string")
	}

	tester.logger.Infof("OK: Signed in username: %s", tester.username)

	return &token, nil
}

func (tester *Tester) validateJWT(token *string) error {

	claims := new(api.JWTClaims)

	_, err := jwt.ParseWithClaims(*token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tester.cfg.JWTSecret), nil
	})
	if err != nil {
		return err
	}

	tester.logger.Infof("OK: JWT valid")

	return nil
}

func (tester *Tester) deactivateTestUser(token *string) error {

	url := fmt.Sprintf("http://%s/api/1.0/auth/deactivate", tester.cfg.Addr)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := tester.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		break
	default:
		return fmt.Errorf("expected status %d in deactivate response, got: %d, response body: %s", http.StatusAccepted, resp.StatusCode, string(body))
	}

	tester.logger.Infof("OK: Deactivated username: %s\n", tester.username)

	return nil
}
