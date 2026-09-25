import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { BaseApiService } from './base-api.service';
import { UserInfo } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class UserService extends BaseApiService<UserInfo, string> {
  protected override get resourceUrl(): string {
    return 'api/users';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  changePassword(data: any): Observable<any> {
    return this.http.post<any>(`${this.url}/change-password`, data);
  }

  getHistory(userId: any): Observable<any[]> {
    return this.http.get<any[]>(`${this.url}/${userId}/history`);
  }

  login(credentials: any): Observable<any> {
    return this.http.post<any>(`${this.url}/login`, credentials);
  }


  verifyToken(authCode: string): Observable<any> {
    return this.http.get<any>(`${this.url.replace('/users', '/auths')}`, {
      params: { authCode }
    });
  }

  logout(): Observable<any> {
    return this.http.post<any>(`${this.url}/logout`, {});
  }

  hasPermission(permission: string): boolean {
    const userStr = localStorage.getItem('currentUser');
    if (!userStr) return false;
    try {
      const user = JSON.parse(userStr);
      const roles: string[] = user.roles || user.Roles || [];
      const isViewer = roles.some(r => r.toLowerCase() === 'viewer');
      if (isViewer) return false;

      const permissions: string[] = user.permissions || user.Permissions || [];
      return permissions.includes(permission);
    } catch {
      return false;
    }
  }
}
