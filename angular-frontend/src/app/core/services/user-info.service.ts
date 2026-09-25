import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';

export interface UserInfoEntity {
  userID: string;
  pass?: string;
  userName?: string;
  description?: string;
  area?: string;
}

@Injectable({
  providedIn: 'root'
})
export class UserInfoService extends BaseApiService<UserInfoEntity, string> {
  protected override get resourceUrl(): string {
    return 'api/users';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
