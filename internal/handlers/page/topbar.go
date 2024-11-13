package topbar

import (
	"html/template"
)

var FuncsTopbar = template.FuncMap{
	"renderAuthButton": renderAuthButton,
}

func renderAuthButton(isLoggedIn bool) string {
	if isLoggedIn {
		return `<a href="/logout" class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Logout</a>`
	}
	return `<a href="/auth/google" class="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Login with Google</a>`
}