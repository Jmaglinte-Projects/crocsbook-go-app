# Authentication flow & frontend tutorial

This app uses **Google Sign-In (backend auth)**: the frontend gets a Google ID token and sends it to the gRPC backend; the backend verifies the token and returns an app JWT. No redirect URL or HTTP callback is required.

---

## 1. Flow overview

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Frontend  │       │   Google    │       │   Backend   │
│   (your app)│       │  (Sign-In)  │       │   (gRPC)    │
└──────┬──────┘       └──────┬──────┘       └──────┬──────┘
       │                     │                     │
       │  1. User clicks     │                     │
       │  "Sign in with      │                     │
       │   Google"           │                     │
       │────────────────────>                     │
       │                     │                     │
       │  2. User signs in   │                     │
       │  (Google popup/     │                     │
       │   One Tap)          │                     │
       │<────────────────────                     │
       │  3. ID token        │                     │
       │                     │                     │
       │  4. gRPC GoogleSignIn(id_token)           │
       │──────────────────────────────────────────>│
       │                     │  5. Verify token    │
       │                     │     (tokeninfo)    │
       │                     │  6. Find/create    │
       │                     │     user, sign JWT │
       │  7. App JWT         │                     │
       │<──────────────────────────────────────────│
       │                     │                     │
       │  8. Use JWT in      │                     │
       │  later gRPC calls   │                     │
       │  (e.g. metadata)    │                     │
       └──────────────────────────────────────────┘
```

**Backend steps (what our service does):**

1. Receive `id_token` from the client.
2. Call Google’s tokeninfo endpoint to validate the token and get claims (`aud`, `iss`, `exp`, `email`, `name`, `picture`).
3. Check `aud` equals our `GOOGLE_OAUTH_CLIENT_ID`.
4. Find user by email; if none, create one.
5. Sign our own JWT (contains user ID), return it.

**Frontend responsibility:** Use Google Sign-In to obtain the ID token, then call our gRPC `GoogleSignIn` with that token and store the returned JWT for authenticated requests.

---

## 2. Backend configuration

Ensure `.env` has:

- **`JWT_SECRET`** – Secret used to sign our JWTs (use a long random string in production).
- **`GOOGLE_OAUTH_CLIENT_ID`** – Google OAuth 2.0 **Web client** Client ID (same one the frontend uses for Google Sign-In).

No redirect URL or client secret is needed for this flow.

---

## 3. Frontend tutorial

### 3.1 Google Cloud setup

1. Open [Google Cloud Console](https://console.cloud.google.com/) → **APIs & Services** → **Credentials**.
2. Create or select an OAuth 2.0 **Web application** client (or use the same one for both web and backend).
3. Copy the **Client ID** and set it in your frontend config and in the backend `.env` as `GOOGLE_OAUTH_CLIENT_ID`.
4. If you use the newer **Google Identity Services (GIS)**:
   - In **APIs & Services** → **Credentials** → your OAuth client, add **Authorized JavaScript origins** (e.g. `http://localhost:5173`, `https://yourdomain.com`).
   - No redirect URI is required for the one-tap / popup flow that returns an ID token to the page.

### 3.2 Load Google Identity Services script

Add the GIS script once (e.g. in `index.html` or your root layout):

```html
<script src="https://accounts.google.com/gsi/client" async defer></script>
```

### 3.3 Sign-in button and get ID token

**Option A – Google One Tap / button (recommended)**

```html
<div id="g_id_onload"
     data-client_id="YOUR_GOOGLE_CLIENT_ID"
     data-callback="onGoogleSignIn"
     data-auto_prompt="false">
</div>
<div class="g_id_signin"
     data-type="standard"
     data-shape="rectangular"
     data-theme="outline"
     data-text="signin_with"
     data-size="large">
</div>
```

```javascript
function onGoogleSignIn(response) {
  // response.credential is the JWT ID token to send to your backend
  const idToken = response.credential;
  callBackendGoogleSignIn(idToken);
}
```

**Option B – Programmatic (e.g. React)**

Use the `google.accounts.id.initialize` and `google.accounts.id.prompt` (or render the button with `google.accounts.id.renderButton`). In the callback you receive a credential object with the ID token:

```javascript
window.onload = function () {
  google.accounts.id.initialize({
    client_id: 'YOUR_GOOGLE_CLIENT_ID',
    callback: (response) => {
      const idToken = response.credential;
      callBackendGoogleSignIn(idToken);
    },
  });
  // One Tap
  google.accounts.id.prompt();
};
```

Replace `YOUR_GOOGLE_CLIENT_ID` with the same Client ID you set in the backend `.env`.

### 3.4 Call the gRPC backend

Send the ID token to the backend via the **GoogleSignIn** gRPC method.

**Request (proto):**

- `GoogleSignInIn.id_token` = the Google ID token string from step 3.3.

**Response (proto):**

- `GoogleSignInOut.token` = your app’s JWT. Store it (e.g. in memory, `localStorage`, or a cookie) and send it on subsequent gRPC calls (e.g. in metadata).

**Example (pseudo-code; exact API depends on your gRPC client):**

```javascript
async function callBackendGoogleSignIn(idToken) {
  const response = await authServiceClient.googleSignIn({
    idToken: idToken,
  });
  const appToken = response.token;
  // Store for later use
  localStorage.setItem('app_token', appToken);
  // Or set in your auth state / context
  setAuthToken(appToken);
}
```

If you use **grpc-web** or **Connect**:

- Create an `AuthService` client from your generated code (pointing at your gRPC-web or Connect gateway URL).
- Call `googleSignIn({ idToken: idToken })` and use the returned `token` as above.

### 3.5 Use the JWT on later requests

Attach the app JWT to authenticated gRPC calls, for example via **metadata** (header):

- Header name is often `Authorization: Bearer <token>` or a custom header your backend expects.
- Set it on every request that requires authentication so the backend can identify the user (e.g. after validating the JWT and reading the user ID from claims).

---

## 4. Summary

| Step | Where | Action |
|------|--------|--------|
| 1 | Google Cloud | Create OAuth Web client, get Client ID. |
| 2 | Backend `.env` | Set `JWT_SECRET` and `GOOGLE_OAUTH_CLIENT_ID`. |
| 3 | Frontend | Load GSI script, add sign-in button or One Tap. |
| 4 | Frontend | In callback, get `response.credential` (ID token). |
| 5 | Frontend | Call gRPC `GoogleSignIn(id_token)` and store returned `token` (JWT). |
| 6 | Frontend | Send JWT on subsequent gRPC calls (e.g. `Authorization: Bearer <token>`). |

No HTTP callback URL is used; the only “callback” is the frontend’s Google Sign-In callback, which then calls your gRPC backend with the ID token.
