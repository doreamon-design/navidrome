import jwtDecode from 'jwt-decode'
// import cookie from '@zodash/cookie';
import { baseUrl } from './utils'
import config from './config'
import { startEventStream, stopEventStream } from './eventStream'

// config sent from server may contain authentication info, for example when the user is authenticated
// by a reverse proxy request header
if (config.auth) {
  try {
    storeAuthenticationInfo(config.auth)
  } catch (e) {
    console.log(e)
  }
}

function storeAuthenticationInfo(authInfo) {
  authInfo.token && localStorage.setItem('token', authInfo.token)
  localStorage.setItem('userId', authInfo.id)
  localStorage.setItem('name', authInfo.name)
  localStorage.setItem('username', authInfo.username)
  authInfo.avatar && localStorage.setItem('avatar', authInfo.avatar)
  localStorage.setItem('role', authInfo.isAdmin ? 'admin' : 'regular')
  localStorage.setItem('subsonic-salt', authInfo.subsonicSalt)
  localStorage.setItem('subsonic-token', authInfo.subsonicToken)
  localStorage.setItem('is-authenticated', 'true')
}

const authProvider = {
  login: ({ username, password }) => {
    let url = baseUrl('/auth/login')
    if (config.firstTime) {
      url = baseUrl('/auth/createAdmin')
    }
    const request = new Request(url, {
      method: 'POST',
      body: JSON.stringify({ username, password }),
      headers: new Headers({ 'Content-Type': 'application/json' }),
    })
    return fetch(request)
      .then((response) => {
        if (response.status < 200 || response.status >= 300) {
          throw new Error(response.statusText)
        }
        return response.json()
      })
      .then((response) => {
        jwtDecode(response.token) // Validate token
        storeAuthenticationInfo(response)
        // Avoid "going to create admin" dialog after logout/login without a refresh
        config.firstTime = false
        if (config.devActivityPanel) {
          startEventStream()
        }
        return response
      })
      .catch((error) => {
        if (
          error.message === 'Failed to fetch' ||
          error.stack === 'TypeError: Failed to fetch'
        ) {
          throw new Error('errors.network_error')
        }

        throw new Error(error)
      })
  },

  logout: () => {
    stopEventStream()
    removeItems()
    // return Promise.resolve()

    window.location.replace('/logout')

    return new Promise((resolve) => {
      setTimeout(() => {
        resolve()
      }, 3000)
    })
  },

  checkAuth: () =>
    localStorage.getItem('is-authenticated') === 'true'
      ? Promise.resolve()
      : Promise.reject(),

  checkError: ({ status }) => {
    if (status === 401) {
      removeItems()
      return Promise.reject()
    }
    return Promise.resolve()
  },

  getPermissions: () => {
    const role = localStorage.getItem('role')
    return role ? Promise.resolve(role) : Promise.reject()
  },

  getIdentity: () => {
    return {
      id: localStorage.getItem('username'),
      fullName: localStorage.getItem('name'),
      avatar: localStorage.getItem('avatar'),
    }
  },

  autoLogin: () => {
    const url = baseUrl('/api/user')

    // const token = cookie.get("go-zoox_oauth2_token")
    // console.log('token:', token);

    const request = new Request(url, {
      method: 'GET',
      headers: new Headers({ Accept: 'application/json' }),
    })
    return fetch(request)
      .then((response) => {
        if (response.status < 200 || response.status >= 300) {
          throw new Error(response.statusText)
        }
        return response.json()
      })
      .then((response) => {
        jwtDecode(response.token) // Validate token
        storeAuthenticationInfo(response)
        // Avoid "going to create admin" dialog after logout/login without a refresh
        config.firstTime = false
        if (config.devActivityPanel) {
          startEventStream()
        }
        return response
      })
      .catch((error) => {
        if (
          error.message === 'Failed to fetch' ||
          error.stack === 'TypeError: Failed to fetch'
        ) {
          throw new Error('errors.network_error')
        }

        // throw new Error(error)
        console.error('failed to get user:', error)
        // authProvider.logout()
      })
  },
}

const removeItems = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('userId')
  localStorage.removeItem('name')
  localStorage.removeItem('username')
  localStorage.removeItem('avatar')
  localStorage.removeItem('role')
  localStorage.removeItem('subsonic-salt')
  localStorage.removeItem('subsonic-token')
  localStorage.removeItem('is-authenticated')
}

export default authProvider
