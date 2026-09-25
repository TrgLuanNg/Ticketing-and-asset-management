import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import { BaseApiService } from './base-api.service';
import { Asset } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class AssetService extends BaseApiService<Asset, string> {
  public assetUpdated$ = new Subject<Asset | string>();

  protected override get resourceUrl(): string {
    return 'api/assets';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  exportExcel(filter: any): Observable<Blob> {
    const options = this.getOptions();
    return this.http.post(`${this.url}/export`, filter, {
      ...options,
      responseType: 'blob'
    });
  }

  importFile(file: File): Observable<any> {
    const formData: FormData = new FormData();
    formData.append('file', file, file.name);
    return this.http.post(`${this.url}/import-file`, formData, this.getOptions());
  }

  searchGrid(filter: any): Observable<any> {
    return this.http.post(`${this.url}/search-grid`, filter, this.getOptions());
  }

  generateUnitCode(type: string, purchasedDate: any, company: string, area: string): Observable<any> {
    const formattedDate = purchasedDate instanceof Date ? purchasedDate.toISOString() : purchasedDate;
    const options = this.getOptions();
    return this.http.get<any>(`${this.url}/generate-unit-code`, {
      ...options,
      params: { type, purchasedDate: formattedDate, company, area }
    });
  }

  downloadTemplate(): Observable<Blob> {
    const options = this.getOptions();
    return this.http.get(`${this.url}/download-template`, {
      ...options,
      responseType: 'blob'
    });
  }

  getDashboardMetrics(fromDate?: string): Observable<any> {
    const options = this.getOptions();
    return this.http.get<any>(`${this.url}/dashboard-metrics`, {
      ...options,
      params: fromDate ? { fromDate } : {}
    });
  }

  getSuggestions(filter: any): Observable<any> {
    const options = this.getOptions();
    return this.http.post<any>(`${this.url}/suggestions`, filter, options);
  }

  getUnassignedSubAssets(filter?: any): Observable<any> {
    const options = this.getOptions();
    return this.http.post<any>(`${this.url}/unassigned-sub-assets`, filter || {}, options);
  }


  updateMainAsset(unitCode: string, mainAsset: string | null): Observable<any> {
    const options = this.getOptions();
    return this.http.put<any>(`${this.url}/${unitCode}/main-asset`, { mainAsset }, options);
  }
}

