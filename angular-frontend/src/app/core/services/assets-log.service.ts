import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';

@Injectable({
  providedIn: 'root'
})
export class AssetsLogService extends BaseApiService<any, number> {
  protected override get resourceUrl(): string {
    return 'api/assetslog';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
