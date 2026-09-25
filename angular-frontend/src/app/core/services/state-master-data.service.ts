import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService, ResponseAPI } from './base-api.service';
import { StateMasterData } from './api.service';
import { Observable } from 'rxjs';
import { shareReplay } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class StateMasterDataService extends BaseApiService<StateMasterData, string> {
  private cache$?: Observable<ResponseAPI<StateMasterData[]>>;

  protected override get resourceUrl(): string {
    return 'api/state-master-datas';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }

  override getAll(): Observable<ResponseAPI<StateMasterData[]>> {
    if (!this.cache$) {
      this.cache$ = super.getAll().pipe(shareReplay(1));
    }
    return this.cache$;
  }

  override create(entity: StateMasterData): Observable<ResponseAPI<StateMasterData>> {
    this.cache$ = undefined;
    return super.create(entity);
  }

  override update(id: string, entity: StateMasterData): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.update(id, entity);
  }

  override delete(id: string): Observable<ResponseAPI<any>> {
    this.cache$ = undefined;
    return super.delete(id);
  }
}
