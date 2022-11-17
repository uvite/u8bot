// @ts-ignore

const Cookies = require('js-cookie');


export   function getCookie(name) {
   return Cookies.get(name)
}

export function setCookie(cookiename, cookievalue) {
    Cookies.set(cookiename, cookievalue, {expires: 365});
}
