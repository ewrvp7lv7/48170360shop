package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/lestrrat-go/jwx/jwt"

	//"github.com/lestrrat-go/jwx/jwt"
	"github.com/EwRvp7LV7/48170360shop/internal/storage/postgres"
)

//InputUserProfile приходящая извне информация для создания пользователя.
type Authentication struct {
	UserName string `json:"account"`
	Password string `json:"password"`
}

func AddRouteAuthentication(router *chi.Mux) {

	router.Use(TokenAuthCtx)
	router.Use(JWTVerifier())

	//пример защищенной страницы
	router.Route("/api/secretpage", func(r chi.Router) {

		r.Use(JWTSecurety) //здесь защита неавторизованного дальше не пустит
		r.Get("/", userAuth)

	})

	router.Route("/api/auth", func(r chi.Router) {

		r.Get("/", checkUserAuthentication)
		r.Post("/", userLogin)
	})

	router.Route("/api/logout", func(r chi.Router) {
		r.Get("/", UserLogout)
	})
}

func userLogin(w http.ResponseWriter, r *http.Request) {

	var data Authentication
	var buf bytes.Buffer

	buf.ReadFrom(r.Body)
	err := json.Unmarshal(buf.Bytes(), &data)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err = validation.Validate(data.Password, validation.Required, validation.RuneLength(6, 15)); err != nil {

		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	jsonbytes, _ := json.Marshal(data)
	uid, passDB, isAdmin, err := postgres.GetUIDPasswordHash(jsonbytes)
	if err != nil {

		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if strings.Compare(strings.TrimSpace(data.Password), passDB) == 0 {
		tokenAuth := r.Context().Value(TokenA{}).(*jwtauth.JWTAuth)
		_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{
			"user_id":   uid,
			"user_name": data.UserName,
			"user_type": isAdmin,
			// "iss":     "",
			// "sub":     "",
			// "aud":     "",
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		})

		//w.Header().Set("Location", "/login")не работает
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", Expires: expiration}
		fmt.Println(cookie)
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
		render.DefaultResponder(w, r, uid)

	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong password!"))
	}
}

func checkUserAuthentication(w http.ResponseWriter, r *http.Request) {

	_, claims, err := jwtauth.FromContext(r.Context())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		fmt.Fprintf(w, `{"logged": false, "user_name": null}`)
		return
	}

	if claims["user_name"] == nil {
		fmt.Println(claims, " ", err)
		fmt.Fprintf(w, `{"logged": false, "user_name": null}`)
		return
	}

	fmt.Fprintf(w, `{"logged": true, "user_name": "%s"}`, claims["user_name"])
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	c := &http.Cookie{
		Name:    "jwt",
		Value:   "clear",
		Path:    "/",
		Expires: time.Now(),

		HttpOnly: true,
	}
	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)
	render.DefaultResponder(w, r, "ok")
}

//example data token
func userAuth(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context()) //можно не проверять err тк JWTSecurety
	w.Write([]byte(fmt.Sprintf(`id_user: %s, user_name: %s, is_manager: %t, `, claims["id_user"], claims["user_name"], claims["user_type"])))
}

//Defense for unaftorisable
func JWTSecurety(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			//http.Error(w, err.Error(), http.StatusUnauthorized)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

//auxiliary: заглушка т.к. ругается что в контекст в качестве ключа вставляем НЕ структуру
type TokenA struct{}

//auxiliary: add tokenAuth to context
func TokenAuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenAuth := jwtauth.New("HS256", []byte("secret key"), nil)
		ctx := context.WithValue(r.Context(), TokenA{}, tokenAuth)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//auxiliary: verify token
func JWTVerifier() func(http.Handler) http.Handler {
	//не нашел способа завести сюда tokenAuth из контекста tokenAuthCtx
	//непонятно почему chi это сделал сразу, а tokenAuth выделен по сути в отдельную глобальную переменную
	tokenAuth := jwtauth.New("HS256", []byte("secret key"), nil)
	return jwtauth.Verifier(tokenAuth)
}

// func AddRouteAuthentication(r *chi.Mux) {

// 	r.Route("/api/secretpage", func(r chi.Router) {

// 		tokenAuth := jwtauth.New("HS256", []byte("secret key"), nil)
// 		r.Use(jwtauth.Verifier(tokenAuth))
// 		r.Use(jwtauth.Authenticator)

// 		r.Get("/", UserAuth)

// 	})

// 	// // Protected routes
// 	// r.Group(func(r chi.Router) {
// 	// 	tokenAuth := jwtauth.New("HS256", []byte("secret key"), nil)
// 	// 	r.Use(jwtauth.Verifier(tokenAuth))
// 	// 	r.Use(jwtauth.Authenticator)

// 	// 	r.Get("/secretpage", func(w http.ResponseWriter, r *http.Request) {
// 	// 		_, claims, _ := jwtauth.FromContext(r.Context())
// 	// 		w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
// 	// 	})

// 	// 	r.Post("/secretpage", func(w http.ResponseWriter, r *http.Request) {
// 	// 		_, claims, _ := jwtauth.FromContext(r.Context())
// 	// 		w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
// 	// 	})
// 	// })

// 	r.Route("/api/auth", func(r chi.Router) {

// 		r.Get("/", UserLoginForm)
// 		r.Post("/", UserLogin)
// 	})

// 	r.Route("/api/logout", func(r chi.Router) {

// 		r.Get("/", UserLogout)

// 	})
// }

// //Хешировать пароли, пока не используются
// func HashPassword(password string) (string, error) {
// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
// 	return string(bytes), err
// }

// func CheckPasswordHash(password, hash string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
// 	return err == nil
// }

// func UserLoginForm(w http.ResponseWriter, r *http.Request) {

// 	form := `<!DOCTYPE html><html><form method="POST" action="/auth">
//   <label>username: <input type="text" name="username" value="username123" autocomplete="username" required></label>
//   <label>password: <input type="password" name="password" value
// ="password123" autocomplete="current-password" required></label>
//   <button type="submit">Login</button>
// </form></html>`
// 	w.Write([]byte(form))
// }

// func UserLogin(w http.ResponseWriter, r *http.Request) {

// 	cur1 := r.FormValue("username")
// 	cur2 := r.FormValue("password")

// 	if cur2 == "password123" {
// 		tokenAuth := jwtauth.New("HS256", []byte("secret key"), nil)
// 		_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{
// 			"name":    cur1,
// 			"user_id": 123,
// 			"iss":     "",
// 			"sub":     "",
// 			"aud":     "",
// 			"exp":     time.Now().Add(15 * time.Minute).Unix(),
// 		})

// 		//w.Header().Set("Location", "/login")не работает
// 		expiration := time.Now().Add(24 * time.Hour)
// 		cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", Expires: expiration}
// 		http.SetCookie(w, &cookie)
// 		fmt.Println("right password!", &cookie)
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("<!DOCTYPE html><html>Post Successful <script type='text/javascript'> document.location = '/secretpage'; </script></html>"))

// 	} else {
// 		UserLoginForm(w, r)
// 	}
// }

// func UserAuth(w http.ResponseWriter, r *http.Request) {
// 	_, claims, _ := jwtauth.FromContext(r.Context())
// 	w.Write([]byte(fmt.Sprintf("Hi user %v!", claims["user_id"])))
// }

// func UserLogout(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("<!DOCTYPE html><html>logout</html>"))
// }
