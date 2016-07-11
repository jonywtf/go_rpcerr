# go_rpcerr
Утилита пробегает по исходникам проетка на GoLang

При обнаружении структур с именем *****ErrorList, например
```Go
type LoginErrorList struct {
	EMAIL_NOT_VERIFICATED       RPCError // CODE:7001 Адрес электронной почты не верифицирован
	INCORRECT_LOGIN_OR_PASSWORD RPCError // CODE:7002 Некорректный логин или пароль
}
```
сгенерирует в отдельном файле <имя оригинального файла>_err.gen.go такой код:
```Go
var LoginError = LoginErrorList{
    EMAIL_NOT_VERIFICATED: RPCError{
        Code:7026,
        Id:"EMAIL_NOT_VERIFICATED",
        Description:"Адрес электронной почты не верифицирован",
        Class:"LoginError",
    },
    INCORRECT_LOGIN_OR_PASSWORD: RPCError{
        Code:7027,
        Id:"INCORRECT_LOGIN_OR_PASSWORD",
        Description:"Некорректный логин или пароль",
        Class:"LoginError",
    },
}
```
Который позволяет удобно возвращать (на фронтенд например) ошибки в виде структур 
```Go
type RPCError struct {
	Id          string `json:"id"`
	Code        uint64 `json:"code"`
	Description string `json:"description"`
	Class       string `json:"class"`
}

func (err RPCError) Error() string ...
```

посредством кода вида:
```Go

func (t *AuthService) Login(ses *session.Session, req *LoginRequest, res *EmptyResponse) RPCError {

	emailData, err := dm.GetEmailByAddress(req.Email)
	if err != nil {
		return LoginError.INCORRECT_LOGIN_OR_PASSWORD
	}
  
  // ....
  
	loginData, err := dm.GetLoginData(emailData.UserId)
	if err != nil {
		return LoginError.INCORRECT_LOGIN_OR_PASSWORD
	}
  
  // ....
  
	return NilError
}
```

При этом объявление, код и описание ошибки записываются в одной строке исходного кода. Это значительно облегчает ведение документации и интеграцию с кодом на других языках программирования посредством json, xml
