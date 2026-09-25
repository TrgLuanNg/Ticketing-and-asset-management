import { HttpInterceptorFn, HttpRequest, HttpHandlerFn, HttpErrorResponse, HttpBackend, HttpClient } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, throwError } from 'rxjs';
import { catchError, filter, take, switchMap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { ResponseAPI } from '../services/base-api.service';

let isRefreshing = false;
const refreshTokenSubject = new BehaviorSubject<string | null>(null);

export const authInterceptor: HttpInterceptorFn = (req: HttpRequest<unknown>, next: HttpHandlerFn): Observable<any> => {
  const router = inject(Router);
  const httpBackend = inject(HttpBackend);

  const isAuthEndpoint = req.url.toLowerCase().includes('/api/auths');

  // Retrieve current user from localStorage
  const userStr = localStorage.getItem('currentUser');
  let token: string | null = null;
  let userId: string | null = null;
  let userName: string | null = null;

  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      token = user.token || user.Token;
      userId = user.userId || user.userID || user.UserId;
      userName = user.userName || user.username || user.UserName;
    } catch (e) {}
  }

  // Clone request and attach headers if not an auth endpoint
  let authReq = req;
  const headersToSet: { [key: string]: string } = {};

  if (!isAuthEndpoint && token && !req.headers.has('Authorization')) {
    headersToSet['Authorization'] = `Bearer ${token}`;
  }
  if (userId && !req.headers.has('X-User-Id')) {
    headersToSet['X-User-Id'] = userId;
  }
  if (userName && !req.headers.has('X-User-Name')) {
    headersToSet['X-User-Name'] = userName;
  }

  if (Object.keys(headersToSet).length > 0) {
    authReq = req.clone({ setHeaders: headersToSet });
  }

  // Send request and handle 401 errors
  return next(authReq).pipe(
    catchError((error: any) => {
      if (error instanceof HttpErrorResponse && error.status === 401) {
        // Immediately log out if auth endpoint itself returned 401
        if (isAuthEndpoint) {
          handleLogout(router);
          return throwError(() => error);
        }

        if (!userStr) {
          handleLogout(router);
          return throwError(() => error);
        }

        try {
          const user = JSON.parse(userStr);
          const refreshToken = user.refreshToken || user.RefreshToken;

          if (!refreshToken) {
            handleLogout(router);
            return throwError(() => error);
          }

          if (!isRefreshing) {
            isRefreshing = true;
            refreshTokenSubject.next(null);

            // Use direct HttpBackend to prevent recursive interceptor execution
            const refreshClient = new HttpClient(httpBackend);

            return refreshClient.post<ResponseAPI<any>>(`${environment.apiUrl}/api/Auths/refresh`, { refreshToken }).pipe(
              switchMap((res: any) => {
                isRefreshing = false;
                if (res && res.status && res.data) {
                  const newAccessToken = res.data.token || res.data.Token;
                  const newRefreshToken = res.data.refreshToken || res.data.RefreshToken;

                  const updatedUser = {
                    ...user,
                    token: newAccessToken,
                    refreshToken: newRefreshToken,
                    permissions: res.data.permissions || res.data.Permissions || user.permissions,
                    roles: res.data.roles || res.data.Roles || user.roles,
                    area: res.data.area || res.data.Area || user.area
                  };

                  localStorage.setItem('currentUser', JSON.stringify(updatedUser));
                  refreshTokenSubject.next(newAccessToken);

                  const updatedUserId = updatedUser.userId || updatedUser.userID || updatedUser.UserId || userId || '';
                  const updatedUserName = updatedUser.userName || updatedUser.username || updatedUser.UserName || userName || '';

                  const retryRequest = authReq.clone({
                    headers: authReq.headers
                      .set('Authorization', `Bearer ${newAccessToken}`)
                      .set('X-User-Id', updatedUserId)
                      .set('X-User-Name', updatedUserName)
                  });

                  return next(retryRequest);
                } else {
                  isRefreshing = false;
                  handleLogout(router);
                  return throwError(() => new Error('Refresh token rotation failed.'));
                }
              }),
              catchError((refreshErr) => {
                isRefreshing = false;
                handleLogout(router);
                return throwError(() => refreshErr);
              })
            );
          } else {
            // Queue concurrent 401 requests while refreshing token
            return refreshTokenSubject.pipe(
              filter((tok): tok is string => tok !== null),
              take(1),
              switchMap((newToken) => {
                const latestUserStr = localStorage.getItem('currentUser');
                let updatedUserId = userId || '';
                let updatedUserName = userName || '';

                if (latestUserStr) {
                  try {
                    const parsed = JSON.parse(latestUserStr);
                    updatedUserId = parsed.userId || parsed.userID || parsed.UserId || updatedUserId;
                    updatedUserName = parsed.userName || parsed.username || parsed.UserName || updatedUserName;
                  } catch (e) {}
                }

                const retryRequest = authReq.clone({
                  headers: authReq.headers
                    .set('Authorization', `Bearer ${newToken}`)
                    .set('X-User-Id', updatedUserId)
                    .set('X-User-Name', updatedUserName)
                });
                return next(retryRequest);
              })
            );
          }
        } catch (e) {
          isRefreshing = false;
          handleLogout(router);
          return throwError(() => error);
        }
      }
      return throwError(() => error);
    })
  );
};

function handleLogout(router: Router) {
  isRefreshing = false;
  localStorage.removeItem('currentUser');
  router.navigate(['/login']);
}
