package server

const embeddedResourceForPasswdchg = `
	IsAuthed: {{.DecodedCookie.IsAuthed}}
	DisplayName: {{.DecodedCookie.DisplayName}}
	ChangeAttemptedAndFailed: {{.Results.ChangeAttemptedAndFailed}}
	ChangeAttemptedAndSucceeded: {{.Results.ChangeAttemptedAndSucceeded}}
	ReasonForFailure: {{.Results.ReasonForFailure}}
`

const embeddedResourceForListusers = `
	IsAuthed: {{.DecodedCookie.IsAuthed}}
	DisplayName: {{.DecodedCookie.DisplayName}}

	Userquery: {{.Results.Userquery}}
	Users: {{.Results.Users}}
	Querysuccess: {{.Results.Querysuccess}}
	Diagnosticmessage: {{.Results.Diagnosticmessage}}
`

func resourceHelper() func() {
	stashedResources := Resources

	Resources = map[string]string{
		"/usermgt/changepasswd.html": embeddedResourceForPasswdchg,
		"/usermgt/listusers.html": embeddedResourceForListusers,
	}

	return func() { Resources = stashedResources }
}
