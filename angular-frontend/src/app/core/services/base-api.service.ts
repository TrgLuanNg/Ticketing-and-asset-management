import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface ResponseAPI<T> {
  status?: boolean;
  message?: string;
  data?: T;
  totalRecords?: number;
}

export interface ModelAdvancedFilter {
  field?: string;
  operator?: string;
  value?: string;
}

export interface Filter {
  pageIndex: number;
  pageSize: number;
  orderBy?: string;
  advancedFilters?: ModelAdvancedFilter[];
}

export abstract class BaseApiService<T, TKey = string | number> {
  protected abstract get resourceUrl(): string;
  protected baseUrl = environment.apiUrl;

  constructor(protected http: HttpClient) {}

  get url(): string {
    return `${this.baseUrl}/${this.resourceUrl}`;
  }

  protected getOptions() {
    const userStr = localStorage.getItem('currentUser');
    let headers = new HttpHeaders();
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        const userId = user.userId || user.userID || user.UserId;
        const userName = user.userName || user.username || user.UserName;
        if (userId) {
          headers = headers.set('X-User-Id', userId);
        }
        if (userName) {
          headers = headers.set('X-User-Name', userName);
        }
      } catch (e) {}
    }
    return { headers };
  }

  search(filter: Filter): Observable<ResponseAPI<T[]>> {
    return this.http.post<ResponseAPI<T[]>>(`${this.url}/search`, filter, this.getOptions());
  }

  getAll(): Observable<ResponseAPI<T[]>> {
    return this.http.get<ResponseAPI<T[]>>(`${this.url}/get-all`, this.getOptions());
  }

  getById(id: TKey): Observable<ResponseAPI<T>> {
    return this.http.get<ResponseAPI<T>>(`${this.url}/${id}`, this.getOptions());
  }

  create(entity: T): Observable<ResponseAPI<T>> {
    return this.http.post<ResponseAPI<T>>(this.url, entity, this.getOptions());
  }

  update(id: TKey, entity: T): Observable<ResponseAPI<any>> {
    return this.http.put<ResponseAPI<any>>(`${this.url}/${id}`, entity, this.getOptions());
  }

  delete(id: TKey): Observable<ResponseAPI<any>> {
    return this.http.delete<ResponseAPI<any>>(`${this.url}/${id}`, this.getOptions());
  }
}
