import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';
import { Handover } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class HandoverService extends BaseApiService<Handover, string> {
  protected override get resourceUrl(): string {
    return 'api/handovers';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
