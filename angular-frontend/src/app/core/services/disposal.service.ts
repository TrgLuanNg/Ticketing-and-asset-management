import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { BaseApiService } from './base-api.service';

@Injectable({
  providedIn: 'root'
})
export class DisposalService extends BaseApiService<any, string> {
  protected override get resourceUrl(): string {
    return 'api/disposals';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  checkUnitCodes(unitCodes: string[]): Observable<any> {
    return this.http.post<any>(`${this.url}/check-unitcodes`, { unitCodes }, this.getOptions());
  }

  transferWfd(unitCodes: string[], note: string, userSession: string): Observable<any> {
    return this.http.post<any>(`${this.url}/transfer-wfd`, { unitCodes, note, userSession }, this.getOptions());
  }

  confirmDispose(idDisposal: string, note: string, userSession: string): Observable<any> {
    return this.http.post<any>(`${this.url}/confirm-dispose`, { idDisposal, note, userSession }, this.getOptions());
  }

  getBatches(): Observable<any> {
    return this.http.get<any>(`${this.url}/batches`, this.getOptions());
  }

  getBatchDetails(idDisposal: string): Observable<any> {
    return this.http.get<any>(`${this.url}/batch-details/${idDisposal}`, this.getOptions());
  }

  importUnitCodes(file: File): Observable<any> {
    const formData = new FormData();
    formData.append('file', file, file.name);
    return this.http.post<any>(`${this.url}/import-unitcodes`, formData, this.getOptions());
  }
}
