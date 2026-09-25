import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService, ResponseAPI } from './base-api.service';
import { Department } from './api.service';
import { Observable } from 'rxjs';
import { shareReplay } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class DepartmentService extends BaseApiService<Department, string> {
  private cache$?: Observable<ResponseAPI<Department[]>>;

  protected override get resourceUrl(): string {
    return 'api/departments';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  override getAll(): Observable<ResponseAPI<Department[]>> {
    if (!this.cache$) {
      this.cache$ = super.getAll().pipe(shareReplay(1));
    }
    return this.cache$;
  }

  override create(entity: Department): Observable<ResponseAPI<Department>> {
    this.cache$ = undefined;
    return super.create(entity);
  }

  override update(id: string, entity: Department): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.update(id, entity);
  }

  override delete(id: string): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.delete(id);
  }
}
