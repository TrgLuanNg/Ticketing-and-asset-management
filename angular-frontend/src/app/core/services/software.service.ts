import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';

export interface Software {
  id?: number;
  unitCode: string;
  software: string;
  version?: string;
}

@Injectable({
  providedIn: 'root'
})
export class SoftwareService extends BaseApiService<Software, number> {
  protected override get resourceUrl(): string {
    return 'api/softwares';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
