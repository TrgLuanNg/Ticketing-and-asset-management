import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';
import { TitleEmployee } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class TitleEmployeeService extends BaseApiService<TitleEmployee, string> {
  protected override get resourceUrl(): string {
    return 'api/title-employees';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
