service for securely storing data and retrieving it later

STATUS CODES
	200 OK
	403 Forbidden
	404 Not Found

TERMS
	runs entirely over HTTP
	CSK client symmetric key for encrypting stored content
	SSK shared symmetric key for securing communications
	SPK server public-private key pair for securing key exchange
	TOKEN unique token encrypted with SSK
	TAG a content identifier encrypted with SSK
	CONTENT text encrypted with CSK
	CSK PER content tag

TOKEN GENERATION
	server
		create new unique token
		assign CONTENT records to new token
		delete old token
		add new token to response

REGISTRATION
	GET /key
	server response
		SPK
	PUT /user
		SSK encrypted with SPK
	server response
		HTTP status
		TOKEN cookie

LOGOUT
	DELETE /user
	server response
		HTTP status

REPLACE KEY
	PUT /TOKEN/key
		CSK encrypted with SPK
	server response
		HTTP status
		TOKEN cookie

STORE CONTENT
	PUT /TOKEN/TAG
		CONTENT
	server response
		HTTP status
		TOKEN cookie

RETRIEVE CONTENT
	GET /TOKEN/TAG
	server response
		HTTP status
		CONTENT
		TOKEN cookie

FORGET CONTENT
	DELETE /TOKEN/TAG
	server response
		HTTP status
		TOKEN cookie

LIST CONTENT
	GET /TOKEN/TAG
	server response
		HTTP status
		CONTENT(s)
		TOKEN cookie
	POST /TOKEN
		TAG(s)
	server response
		HTTP status
		CONTENT(s)
		TOKEN cookie

EXPIRE KEY
	server deletes SSK
	client REPLACE KEY