import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService, ResponseAPI } from './base-api.service';
import { AssetType } from './api.service';
import { Observable, of } from 'rxjs';
import { catchError, shareReplay } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class TypeService extends BaseApiService<AssetType, string> {
  private cache$?: Observable<ResponseAPI<AssetType[]>>;
  private deviceConfigCache = new Map<string, Observable<any>>();

  protected override get resourceUrl(): string {
    return 'api/types';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  override getAll(): Observable<ResponseAPI<AssetType[]>> {
    if (!this.cache$) {
      this.cache$ = super.getAll().pipe(shareReplay(1));
    }
    return this.cache$;
  }

  override create(entity: AssetType): Observable<ResponseAPI<AssetType>> {
    this.cache$ = undefined;
    return super.create(entity);
  }

  override update(id: string, entity: AssetType): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.update(id, entity);
  }

  override delete(id: string): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.delete(id);
  }

  getDeviceConfig(type: string): Observable<any> {
    const key = String(type).trim().toUpperCase();
    if (!this.deviceConfigCache.has(key)) {
      const req$ = this.http.get<any>(`${this.baseUrl}/api/DeviceConfigs/${type}`).pipe(
        catchError(() => of({ status: false, data: null, message: 'Device config not found' })),
        shareReplay(1)
      );
      this.deviceConfigCache.set(key, req$);
    }
    return this.deviceConfigCache.get(key)!;
  }

  clearDeviceConfigCache(): void {
    this.deviceConfigCache.clear();
  }

  saveDeviceConfig(type: string, payload: any): Observable<any> {
    this.clearDeviceConfigCache();
    return this.http.put<any>(`${this.baseUrl}/api/DeviceConfigs/${type}`, payload);
  }

  createDeviceConfig(payload: any): Observable<any> {
    this.clearDeviceConfigCache();
    return this.http.post<any>(`${this.baseUrl}/api/DeviceConfigs`, payload);
  }

  deleteDeviceConfig(type: string): Observable<any> {
    this.clearDeviceConfigCache();
    return this.http.delete<any>(`${this.baseUrl}/api/DeviceConfigs/${type}`);
  }
}
