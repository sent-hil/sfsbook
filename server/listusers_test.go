package server

import (
	"context"
//	"fmt"
	"io/ioutil"
//	"net/http"
	"net/http/httptest"
//	"reflect"
//	"strings"
	"testing"

//	"github.com/pborman/uuid"
	"github.com/rjkroege/mocking"
//	"golang.org/x/crypto/bcrypt"
)

func makeUnderTestHandlerListUsers(tape *mocking.Tape) *listUsers {
	undertesthandler := &listUsers{
		// Always use the embedded resource.
		embr:         makeEmbeddableResource(""),
		passwordfile: (*mockPasswordIndex)(tape),
	}
	return undertesthandler
}

func TestListUsesNotsignedIn(t *testing.T) {
	defer resourceHelper()()

	undertesthandler := makeUnderTestHandlerListUsers(nil)

	testreq := httptest.NewRequest("GET", "https://sfsbook.org/usermgt/listusers.html", nil)
	recorder := httptest.NewRecorder()
	testreq = testreq.WithContext(context.WithValue(testreq.Context(), UserCookieStateName, new(UserCookie)))

	undertesthandler.ServeHTTP(recorder, testreq)

	result := recorder.Result()
	if got, want := result.StatusCode, 200; got != want {
		t.Errorf("bad response code: got %v, want %v", got, want)
	}
	resultAsString, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatal("couldn't read recorded response", err)
	}

	if got, want := string(resultAsString), "\n\tIsAuthed: false\n\tDisplayName: \n\n\tUserquery: \n\tUsers: []\n\tQuerysuccess: false\n\tDiagnosticmessage: \n"; got != want {
		t.Errorf("bad response body: got %v\n(%#v)\nwant %v\n(%#v)", got, got, want, want)
	}
}
