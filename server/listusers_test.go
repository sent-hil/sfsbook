package server

import (
	"context"
//	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
//	"strings"
	"testing"

	"github.com/pborman/uuid"
	"github.com/rjkroege/mocking"
//	"golang.org/x/crypto/bcrypt"
)

// makeUnderTestHandlerListUsers creates a listUsers structure that
// uses a mock (tape based) implementation of PasswordIndex.
func makeUnderTestHandlerListUsers(tape *mocking.Tape) *listUsers {
	undertesthandler := &listUsers{
		// Always use the embedded resource.
		embr:         makeEmbeddableResource(""),
		passwordfile: (*mockPasswordIndex)(tape),
	}
	return undertesthandler
}

// TestListUsersNotsignedIn shows that the server does not
// let requests with invalid cookies retrieve the user list.
func TestListUsersNotsignedIn(t *testing.T) {
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


// TestListUsersSignedInNoAdmin shows that a user with a valid cookie but no capability
// to list users is not permitted to do so.
func TestListUsersSignedInNoAdmin(t *testing.T) {
	uuid := uuid.NewRandom()
	defer resourceHelper()()
	undertesthandler := makeUnderTestHandlerListUsers(nil)

	testreq := httptest.NewRequest("GET", "https://sfsbook.org/usermgt/listusers.html", nil)
	recorder := httptest.NewRecorder()

	// User does not have the right to view users.
	usercookie := &UserCookie{
		Uuid:        uuid,
		Capability:  CapabilityViewPublicResourceEntry | CapabilityViewOwnVolunteerComment | CapabilityViewOtherVolunteerComment | CapabilityEditOwnVolunteerComment | CapabilityEditOtherVolunteerComment | CapabilityEditResource | CapabilityInviteNewVolunteer | CapabilityInviteNewAdmin,
		Displayname: "Homer Simpson",
		// Time not needed.
	}
	testreq = testreq.WithContext(context.WithValue(testreq.Context(), UserCookieStateName, usercookie))

	// Run handler.
	undertesthandler.ServeHTTP(recorder, testreq)

	// Expect that the user is not allowed to see users.
	// something is wrong here!
	result := recorder.Result()
	if got, want := result.StatusCode, 200; got != want {
		t.Errorf("bad response code: got %v, want %v", got, want)
	}
	resultAsString, err := ioutil.ReadAll(result.Body)
	if err != nil {
		t.Fatal("couldn't read recorded response", err)
	}

	if got, want := string(resultAsString), "\n\tIsAuthed: true\n\tDisplayName: Homer Simpson\n\n\tUserquery: \n\tUsers: []\n\tQuerysuccess: false\n\tDiagnosticmessage: Sign in as an admin to list users.\n"; got != want {
		t.Errorf("bad response body: got %v\n(%#v)\nwant %v\n(%#v)", got, got, want, want)
	}
}

func addCookie(req *http.Request) *http.Request {
	uuid := uuid.NewRandom()
	// User does have the capability to view users.
	usercookie := &UserCookie{
		Uuid:        uuid,
		Capability:  CapabilityViewUsers ,
		Displayname: "Homer Simpson",
	}
	return req.WithContext(context.WithValue(req.Context(),
		UserCookieStateName, usercookie))
}

type testPattern struct {
	urlargs string
	tapeResponse interface{}
	tapeRecord []interface{}
	outputString string	
}

// TestListUsersShowBasicList shows that a user with capability can list
// the currently configured users.
func TestListUsersShowBasicList(t *testing.T) {
	defer resourceHelper()()
	tape := mocking.NewTape()
	undertesthandler := makeUnderTestHandlerListUsers(tape)

	testPatterns := []testPattern{
		{
			"",
			[]map[string]interface{}{
				{
					"display_name": "Homer Simpson",
				},
				{
					"display_name": "Lisa Simpson",
				},
			},
			[]interface {}{listUsersStim{query:"", size:10, from:0}},
			"\n\tIsAuthed: true\n\tDisplayName: Homer Simpson\n\n\tUserquery: \n\tUsers: [map[display_name:Homer Simpson] map[display_name:Lisa Simpson]]\n\tQuerysuccess: true\n\tDiagnosticmessage: \n",
		},
		{
			"",
			[]map[string]interface{}{},
			[]interface {}{listUsersStim{query:"", size:10, from:0}},
			"\n\tIsAuthed: true\n\tDisplayName: Homer Simpson\n\n\tUserquery: \n\tUsers: []\n\tQuerysuccess: false\n\tDiagnosticmessage: Userquery matches no users.\n",
		},
		{
			"?userquery=pandabear",
			[]map[string]interface{}{
				{
					"display_name": "Homer Simpson",
				},
			},
			[]interface {}{listUsersStim{query:"pandabear", size:10, from:0}},
			"\n\tIsAuthed: true\n\tDisplayName: Homer Simpson\n\n\tUserquery: pandabear\n\tUsers: [map[display_name:Homer Simpson]]\n\tQuerysuccess: true\n\tDiagnosticmessage: \n",
		},

		// start of some more interesting tests
		// this is a URL string that changes the selected to a 
		// /usermgt/listusers.html
// I'm really surprised that this passes. What's the deal with this anyway?
		{
			"?userquery=&selected-0=lisasimpson&rolechange=admin",
			[]map[string]interface{}{
				{
					"display_name": "Homer Simpson",
				},
				{
					"display_name": "Lisa Simpson",
				},
			},
			[]interface {}{listUsersStim{query:"", size:10, from:0}},
			"\n\tIsAuthed: true\n\tDisplayName: Homer Simpson\n\n\tUserquery: \n\tUsers: [map[display_name:Homer Simpson] map[display_name:Lisa Simpson]]\n\tQuerysuccess: true\n\tDiagnosticmessage: Showing all...\n",
		},
	}

	for _, tp  := range testPatterns {
		testreq := httptest.NewRequest("GET",
			"https://sfsbook.org/usermgt/listusers.html" + tp.urlargs, nil)
		recorder := httptest.NewRecorder()
		testreq = addCookie(testreq)

		// Note simplified user data to avoid the issue that the maps are not
		// emitted in a consistent order.
		tape.Rewind()
		tape.SetResponses(tp.tapeResponse)

		// Run handler.
		undertesthandler.ServeHTTP(recorder, testreq)

		result := recorder.Result()
		if got, want := result.StatusCode, 200; got != want {
			t.Errorf("bad response code: got %v, want %v", got, want)
		}
		resultAsString, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatal("couldn't read recorded response", err)
		}

		if got, want := string(resultAsString), tp.outputString; got != want {
			t.Errorf("bad response body: got %v\n(%#v)\nwant %v\n(%#v)", got, got, want, want)
		}

		playedtape := tape.Play()
		if got, want := playedtape, tp.tapeRecord; !reflect.DeepEqual(want, got) {
			t.Errorf("invalid call sequence. Got %#v, want %#v", got, want)
		}
	}
}
