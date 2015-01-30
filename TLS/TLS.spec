service for securely storing data and retrieving it later

ENCODINGS
	identifiers passed in URLs are Base32

STATUS CODES
	200 OK
	403 Forbidden
	404 Not Found

TERMS
	runs entirely over HTTPS
	CSK client symmetric key for encrypting stored content
	TOKEN a unique token identifying user
	FILE text encrypted with CSK
	TAG a FILE NAME encrypted with CSK

STATUS
	GET /
	server response
		HTTP status
		server stats

REGISTRATION
	GET /user
	server response
		HTTP status
		TOKEN

USER STATUS
	GET /user/TOKEN
	server response
		HTTP status

FORGET USER
	DELETE /user/TOKEN
	server response
		HTTP status

STORE FILE
	PUT /file/TOKEN/TAG
		FILE
	server response
		HTTP status

RETRIEVE FILE
	GET /file/TOKEN/TAG
	server response
		HTTP status
		FILE

FORGET FILE
	DELETE /file/TOKEN/TAG
	server response
		HTTP status

LIST FILES
	GET /file/TOKEN/
	server response
		HTTP status
		FILE(s)