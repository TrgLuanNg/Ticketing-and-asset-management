import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, shareReplay } from 'rxjs/operators';
import { BaseApiService } from './base-api.service';
import { TechSpecKey } from './api.service';

@Injectable({
  providedIn: 'root'
})
export class TechSpecService extends BaseApiService<TechSpecKey, string> {
  private itemGroupCache = new Map<string, Observable<any>>();

  protected override get resourceUrl(): string {
    return 'api/techspecs';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  // Master: Tech Spec Groups
  getGroups(): Observable<any[]> {
    return this.http.get<any[]>(this.url);
  }

  createGroup(group: any): Observable<any> {
    return this.http.post<any>(this.url, group);
  }

  // Detail: Items for a group (Cached to eliminate duplicate HTTP requests)
  getItemsByGroup(groupId: any): Observable<any> {
    if (!groupId || groupId === 'NONE') {
      return of({ status: true, data: [] });
    }

    const key = String(groupId).trim().toUpperCase();
    if (!this.itemGroupCache.has(key)) {
      const filter = {
        pageIndex: 1,
        pageSize: 1000,
        advancedFilters: [
          { field: 'Item', operator: '==', value: groupId }
        ]
      };
      const req$ = this.http.post<any>(`${this.baseUrl}/api/item-master-datas/search`, filter).pipe(
        shareReplay(1)
      );
      this.itemGroupCache.set(key, req$);
    }
    return this.itemGroupCache.get(key)!;
  }

  clearGroupCache(): void {
    this.itemGroupCache.clear();
  }

  createItem(groupId: any, item: any): Observable<any> {
    this.clearGroupCache();
    return this.http.post<any>(`${this.baseUrl}/api/item-master-datas`, item);
  }

  updateItem(groupId: any, itemId: any, item: any): Observable<any> {
    this.clearGroupCache();
    return this.http.put<any>(`${this.baseUrl}/api/item-master-datas/${itemId}`, item);
  }

  deleteItem(groupId: any, itemId: any): Observable<any> {
    this.clearGroupCache();
    return this.http.delete<any>(`${this.baseUrl}/api/item-master-datas/${itemId}`);
  }
}
