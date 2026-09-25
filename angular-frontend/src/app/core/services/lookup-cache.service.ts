import { Injectable, signal } from '@angular/core';
import { DepartmentService } from './department.service';
import { StateMasterDataService } from './state-master-data.service';
import { TypeService } from './type.service';
import { Observable, forkJoin, of } from 'rxjs';
import { map } from 'rxjs/operators';

@Injectable({ providedIn: 'root' })
export class LookupCacheService {
  readonly orgRawList = signal<any[]>([]);
  readonly departmentList = signal<any[]>([]);
  readonly filterDepartmentList = signal<any[]>([]);
  readonly officeList = signal<any[]>([]);
  readonly companyList = signal<any[]>([]);
  readonly vendorList = signal<any[]>([]);
  readonly divisionList = signal<any[]>([]);
  readonly filterDivisionList = signal<any[]>([]);
  readonly groupDepList = signal<any[]>([]);
  readonly stateList = signal<any[]>([]);
  readonly typeList = signal<any[]>([]);
  readonly loaded = signal(false);

  private loadingObservable$: Observable<boolean> | null = null;

  constructor(
    private deptService: DepartmentService,
    private stateService: StateMasterDataService,
    private typeService: TypeService,
  ) {}

  ensureLoaded(): Observable<boolean> {
    if (this.loaded()) {
      return of(true);
    }
    if (this.loadingObservable$) {
      return this.loadingObservable$;
    }

    this.loadingObservable$ = forkJoin({
      depts: this.deptService.getAll(),
      states: this.stateService.getAll(),
      types: this.typeService.getAll()
    }).pipe(
      map(({ depts, states, types }: any) => {
        // Organization data
        const rawDept = depts.status && depts.data ? depts.data : (Array.isArray(depts) ? depts : []);
        this.orgRawList.set(rawDept);

        const deptList = rawDept.filter((item: any) => (item.department === true && item.parentId) || item.id === '00000');
        const noDept = deptList.find((x: any) => x.id === '00000');
        if (noDept) {
          noDept.name = '0';
        } else {
          deptList.push({ id: '00000', name: '0', department: true });
        }
        this.departmentList.set(deptList);

        const filterDept = rawDept.filter((item: any) => item.department === true || item.id === '00000');
        const noDeptFilter = filterDept.find((x: any) => x.id === '00000');
        if (noDeptFilter) {
          noDeptFilter.name = '0';
        } else {
          filterDept.push({ id: '00000', name: '0', department: true });
        }
        this.filterDepartmentList.set(filterDept);

        this.officeList.set(rawDept.filter((item: any) => item.office === true));
        this.companyList.set(rawDept.filter((item: any) => item.company === true));
        this.vendorList.set(rawDept.filter((item: any) => item.vendor === true));
        this.divisionList.set(rawDept.filter((item: any) => item.division === true && item.parentId));
        this.filterDivisionList.set(rawDept.filter((item: any) => item.division === true));
        this.groupDepList.set(rawDept.filter((item: any) => item.groupDivision === true));

        // States data
        const rawStates = states.status && states.data ? states.data : (Array.isArray(states) ? states : []);
        this.stateList.set(rawStates);

        // Types data
        const rawTypes = types.status && types.data ? types.data : (Array.isArray(types) ? types : []);
        const sortedTypes = [...rawTypes].sort((a: any, b: any) => {
          const typeA = (a.type || '').toUpperCase();
          const typeB = (b.type || '').toUpperCase();
          if (typeA === 'LTP' && typeB !== 'LTP') return -1;
          if (typeB === 'LTP' && typeA !== 'LTP') return 1;
          if (typeA === 'DPC' && typeB !== 'DPC') return -1;
          if (typeB === 'DPC' && typeA !== 'DPC') return 1;
          return 0;
        });
        this.typeList.set(sortedTypes);

        this.loaded.set(true);
        this.loadingObservable$ = null;
        return true;
      })
    );

    return this.loadingObservable$;
  }
}
