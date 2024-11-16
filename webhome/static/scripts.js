function setCookie(cname,cvalue,exdays) {
  const d = new Date();
  d.setTime(d.getTime() + (exdays*24*60*60*1000));
  let expires = "expires=" + d.toUTCString();
  document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(cname) {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');
  for(let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) === 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

function check_theme() {
    let theme = getCookie("light_theme");
    if (theme === "light") {
        document.getElementById("csstheme").href = "/static/style_light.css";
    } else {
        document.getElementById("csstheme").href = "/static/style_dark.css";
    }
}

function toggle_theme() {
    let theme = getCookie("light_theme");
    if (theme === "light") {
        setCookie("light_theme", "dark", 30);
    } else {
        setCookie("light_theme", "light", 30);
    }
    check_theme();
}