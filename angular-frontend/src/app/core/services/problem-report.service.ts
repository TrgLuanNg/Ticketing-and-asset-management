import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';
import { ProblemReport } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class ProblemReportService extends BaseApiService<ProblemReport, string> {
  protected override get resourceUrl(): string {
    return 'api/problemreports';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
