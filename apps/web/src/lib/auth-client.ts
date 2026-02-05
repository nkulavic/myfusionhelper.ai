import {
  signIn,
  signUp,
  confirmSignUp,
  signOut,
  fetchAuthSession,
  getCurrentUser,
} from 'aws-amplify/auth'

export async function cognitoSignIn(email: string, password: string) {
  return signIn({ username: email, password })
}

export async function cognitoSignUp(email: string, password: string, name: string) {
  return signUp({
    username: email,
    password,
    options: {
      userAttributes: {
        email,
        name,
      },
    },
  })
}

export async function cognitoConfirmSignUp(email: string, code: string) {
  return confirmSignUp({ username: email, confirmationCode: code })
}

export async function cognitoSignOut() {
  return signOut()
}

export async function getSession() {
  return fetchAuthSession()
}

export { getCurrentUser }
