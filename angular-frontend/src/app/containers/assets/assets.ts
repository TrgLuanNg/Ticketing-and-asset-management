import { Component, OnInit, ChangeDetectorRef, ViewChild, ElementRef, HostListener, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzDatePickerModule } from 'ng-zorro-antd/date-picker';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzStepsModule } from 'ng-zorro-antd/steps';
import { NzAutocompleteModule } from 'ng-zorro-antd/auto-complete';
import { NzPaginationModule } from 'ng-zorro-antd/pagination';
import { Asset } from '../../core/services/api.service';
import { AssetService } from '../../core/services/asset.service';
import { HandoverService } from '../../core/services/handover.service';
import { ProblemReportService } from '../../core/services/problem-report.service';
import { AssetsLogService } from '../../core/services/assets-log.service';
import { DepartmentService } from '../../core/services/department.service';
import { StateMasterDataService } from '../../core/services/state-master-data.service';
import { TypeService } from '../../core/services/type.service';
import { TechSpecService } from '../../core/services/tech-spec.service';
import { DisposalService } from '../../core/services/disposal.service';
import { UserService } from '../../core/services/user.service';
import { LookupCacheService } from '../../core/services/lookup-cache.service';
import saveAs from 'file-saver';
import { ActivatedRoute, Router } from '@angular/router';
import { EMPTY, Subject } from 'rxjs';
import { tap, debounceTime, distinctUntilChanged, switchMap } from 'rxjs/operators';

@Component({
  selector: 'app-assets',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzButtonModule,
    NzIconModule,
    NzTabsModule,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzDatePickerModule,
    NzGridModule,
    NzModalModule,
    NzTableModule,
    NzDropDownModule,
    NzCheckboxModule,
    NzCollapseModule,
    NzStepsModule,
    NzAutocompleteModule
  ],
  templateUrl: './assets.html',
  styleUrl: './assets.less',
  providers: [NzModalService],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Assets implements OnInit {
  @ViewChild('fileInput') fileInput!: ElementRef;
  @ViewChild('stepsWrapper') stepsWrapper!: ElementRef;

  public rowData: any[] = [];
  public searchText: string = '';
  public suggestions: any[] = [];
  public loading = false;
  
  public handoverRowData: any[] = [];

  public problemRowData: any[] = [];

  public logRowData: any[] = [];

  public departmentList: any[] = [];
  public filterDepartmentList: any[] = [];
  public officeList: any[] = [];
  public companyList: any[] = [];
  public vendorList: any[] = [];
  public stateList: any[] = [];
  public divisionList: any[] = [];
  public filterDivisionList: any[] = [];
  public groupDepList: any[] = [];

  public selectedFilters: any[] = [];
  public filterFields: any[] = [];

  public totalRecords: number = 0;

  selectedAsset: any = null;
  selectedHandover: any = null;
  isFilterVisible = false;
  currentFilter: any = {
    pageIndex: 1,
    pageSize: 20,
    orderBy: 'PurchasedDate desc',
    advancedFilters: [
      { field: 'Type', operator: '==', value: 'DPC/LTP/dpc/ltp' },
      { field: 'State', operator: '==', value: '001/002/003' }
    ]
  };
  
  assetForm!: FormGroup;
  subAssetForm!: FormGroup;
  handoverForm!: FormGroup;
  problemForm!: FormGroup;
  specForm!: FormGroup;
  filterForm!: FormGroup;
  addForm!: FormGroup;

  get formVal() {
    return this.addForm?.value || {};
  }

  get assetFormVal() {
    return this.assetForm?.value || {};
  }

  get subAssetFormVal() {
    return this.subAssetForm?.value || {};
  }

  get handoverFormVal() {
    return this.handoverForm?.value || {};
  }

  get problemFormVal() {
    return this.problemForm?.value || {};
  }

  get specFormVal() {
    return this.specForm?.value || {};
  }

  get filterFormVal() {
    return this.filterForm?.value || {};
  }

  isAddModalVisible = false;
  addActiveDeviceConfig: any = null;
  addDetailOptions: { [key: string]: any[] } = {};
  orgRawList: any[] = [];

  typeList: any[] = [];
  activeSpecTarget: string = 'main';
  activeDeviceConfig: any = null;
  detailOptions: { [key: string]: any[] } = {};

  isPrintVisible = false;
  isHandoverVisible = false;
  isProblemVisible = false;
  isEditingDetails = false;
  currentAddStep = 0;
  activeAddSpecTarget: string = 'main';
  selectedExtraSubAssetType: string | null = null;
  extraSubAssetTypes: string[] = [];
  availableSubAssetTypes: any[] = [];
  selectableExtraSubAssetTypes: any[] = [];
  subAssetConfigsData: { [typeCode: string]: { config: any, detailOptions: any } } = {};
  activeDetailTabIndex = 0;
  printData: any = {};
  logFromDate: Date | null = null;
  logToDate: Date | null = null;

  isTargetFieldsRequired = false;

  isDisposalModalVisible = false;
  disposalBatches: any[] = [];
  selectedDisposalBatch: any = null;
  disposalBatchDetails: any[] = [];
  disposalInputCodes: string = '';
  disposalCheckedAssets: any[] = [];
  disposalWfdNote: string = '';
  disposalConfirmNote: string = '';
  selectedProblem: any = null;
  problemEditMode = false;
  isDisposalTransferEnabled = false;
  allowedAreas: string[] = [];
  private searchSubject = new Subject<string>();

  constructor(
    private assetService: AssetService,
    private handoverService: HandoverService,
    private problemReportService: ProblemReportService,
    private departmentService: DepartmentService,
    private stateMasterDataService: StateMasterDataService,
    private assetsLogService: AssetsLogService,
    private typeService: TypeService,
    private techSpecService: TechSpecService,
    private disposalService: DisposalService,
    private cdr: ChangeDetectorRef,
    private message: NzMessageService,
    private modalService: NzModalService,
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private userService: UserService,
    private lookupCacheService: LookupCacheService
  ) {
    this.createForms();
  }

  ngOnInit() {
    this.initializeDefaultFilters();

    this.searchSubject.pipe(
      debounceTime(400),
      distinctUntilChanged(),
      switchMap(value => {
        if (!value || value.trim() === '') {
          this.suggestions = [];
          this.currentFilter.searchText = '';
          this.currentFilter.pageIndex = 1;
          this.loadAssets();
          this.cdr.markForCheck();
          return EMPTY;
        }
        this.currentFilter.searchText = value;
        this.currentFilter.pageIndex = 1;
        this.loadAssets();
        const filterCopy = JSON.parse(JSON.stringify(this.currentFilter));
        return this.assetService.getSuggestions(filterCopy).pipe(
          tap(res => {
            this.suggestions = res.status && res.data ? res.data : [];
            this.cdr.markForCheck();
          })
        );
      })
    ).subscribe();

    const user = this.getCurrentUser();
    this.allowedAreas = [];

    if (user) {
      const currentUserId = user.userId || user.userID || user.UserId;
      if (currentUserId) {
        this.userService.getById(currentUserId).subscribe({
          next: (res: any) => {
            if (res && res.status && res.data) {
              const freshArea = res.data.area ?? res.data.Area ?? '';
              const freshRoles = res.data.roles ?? res.data.Roles ?? user.roles ?? user.Roles;
              const freshPerms = res.data.permissions ?? res.data.Permissions ?? user.permissions ?? user.Permissions;
              const updatedUser = {
                ...user,
                area: freshArea,
                roles: freshRoles,
                permissions: freshPerms
              };
              localStorage.setItem('currentUser', JSON.stringify(updatedUser));
              this.allowedAreas = (freshArea as string).split(',')
                .map((a: string) => a.trim())
                .map((a: string) => a.charAt(0).toUpperCase() + a.slice(1).toLowerCase())
                .filter((a: string) => a.length > 0 && a.toUpperCase() !== 'VN');
              this.initializeFilterFields();
              this.cdr.detectChanges();
            }
          }
        });
      }
    }

    this.route.queryParams.subscribe(params => {
      const action = params['action'];
      if (action) {
        this.router.navigate([], { queryParams: { action: null }, queryParamsHandling: 'merge' });
        setTimeout(() => {
          if (action === 'add') {
            this.onAdd();
          } else if (action === 'disposal') {
            this.onDisposal();
          } else if (action === 'import') {
            this.onImport();
          } else if (action === 'export') {
          }
        }, 300);
      }

      if (params['state']) {
        this.filterForm.patchValue({ state: params['state'] });
        this.submitAdvancedFilter();
      } else if (params['purchasedFrom']) {
        const parts = params['purchasedFrom'].split('/');
        if (parts.length === 2) {
          const formattedDate = `${parts[1]}-${parts[0]}-01`;
          this.currentFilter = {
            pageIndex: 1,
            pageSize: 20,
            orderBy: 'PurchasedDate desc',
            advancedFilters: [
              { field: 'Type', operator: '==', value: 'DPC/LTP/dpc/ltp' },
              { field: 'PurchasedDate', operator: '>=', value: formattedDate }
            ]
          };
          this.loadAssets();
        }
      } else {
        if (this.isViewer()) {
          this.currentFilter = {
            pageIndex: 1,
            pageSize: 20,
            orderBy: 'PurchasedDate desc',
            advancedFilters: []
          };
        } else {
          this.currentFilter = {
            pageIndex: 1,
            pageSize: 20,
            orderBy: 'PurchasedDate desc',
            advancedFilters: [
              { field: 'Type', operator: '==', value: 'DPC/LTP/dpc/ltp' },
              { field: 'State', operator: '==', value: '001/002/003' }
            ]
          };
        }
        this.loadAssets();
      }
    });
    this.loadMetadata();

    this.assetService.assetUpdated$.subscribe((updated: any) => {
      if (typeof updated === 'object' && updated && updated.unitCode) {
        const idx = this.rowData.findIndex(x => x.unitCode === updated.unitCode);
        if (idx !== -1) {
          this.rowData[idx] = { ...this.rowData[idx], ...updated };
          this.rowData = [...this.rowData];
          this.cdr.markForCheck();
        }
      } else if (typeof updated === 'string') {
        this.refreshSingleAssetRow(updated);
      }
    });
  }

  loadMetadata() {
    this.lookupCacheService.ensureLoaded().subscribe(() => {
      this.orgRawList = this.lookupCacheService.orgRawList();
      this.departmentList = this.lookupCacheService.departmentList();
      this.filterDepartmentList = this.lookupCacheService.filterDepartmentList();
      this.officeList = this.lookupCacheService.officeList();
      this.companyList = this.lookupCacheService.companyList();
      this.vendorList = this.lookupCacheService.vendorList();
      this.divisionList = this.lookupCacheService.divisionList();
      this.filterDivisionList = this.lookupCacheService.filterDivisionList();
      this.groupDepList = this.lookupCacheService.groupDepList();
      this.stateList = this.lookupCacheService.stateList();
      this.typeList = this.lookupCacheService.typeList();
      this.initializeFilterFields();
      this.updateSubAssetLists();
      this.cdr.markForCheck();
    });
  }

  updateSubAssetLists() {
    const defaultCodes = ['MON', 'MNT', 'UPS', 'MOU', 'KBO'];
    const activeCodes = [...defaultCodes, ...this.extraSubAssetTypes];
    const currentMainType = this.addForm?.get('type')?.value?.toUpperCase() || '';
    this.availableSubAssetTypes = this.typeList.filter(t => activeCodes.includes(t.type?.toUpperCase()) && t.type?.toUpperCase() !== currentMainType);
    this.selectableExtraSubAssetTypes = this.typeList.filter(t => !activeCodes.includes(t.type?.toUpperCase()) && t.type?.toUpperCase() !== currentMainType);
  }

  resetSubAssetStates() {
    this.extraSubAssetTypes = [];
    this.selectedExtraSubAssetType = null;
    this.activeAddSpecTarget = 'main';
    this.updateSubAssetLists();
  }

  trackByType(index: number, item: any): string {
    return item.type || index;
  }

  get totalAssets(): number {
    return this.rowData.length;
  }

  get activeAssets(): number {
    return this.rowData.filter(x => x.status === 'Active').length;
  }

  get maintenanceAssets(): number {
    return this.rowData.filter(x => x.status === 'Service' || x.status === 'Maintenance' || x.status === 'Repair').length;
  }

  get retiredAssets(): number {
    return this.rowData.filter(x => x.status === 'Retired' || x.status === 'Disposal' || x.status === 'Liquidated').length;
  }

  get filteredRowData(): any[] {
    return this.rowData;
  }

  onSearchChange(value: string): void {
    this.searchText = value;
    this.searchSubject.next(value);
  }

  onSearch(value?: string): void {
    if (value !== undefined) {
      this.searchText = value;
    }
    const text = (this.searchText || '').trim();
    this.currentFilter.searchText = text;
    this.currentFilter.pageIndex = 1;
    this.loadAssets();

    if (!text) {
      this.suggestions = [];
    } else {
      const filterCopy = JSON.parse(JSON.stringify(this.currentFilter));
      this.assetService.getSuggestions(filterCopy).subscribe({
        next: (res: any) => {
          this.suggestions = res.status && res.data ? res.data : [];
          this.cdr.markForCheck();
        }
      });
    }
    this.cdr.markForCheck();
  }

  onPageIndexChange(index: number) {
    this.currentFilter.pageIndex = index;
    this.loadAssets(true);
  }

  onPageSizeChange(size: number) {
    this.currentFilter.pageSize = size;
    this.currentFilter.pageIndex = 1;
    this.loadAssets(false);
  }

  createForms() {
    this.assetForm = this.fb.group({
      unitCode: [null, [Validators.required]],
      type: [null, [Validators.required]],
      department: [null],
      userName: [null],
      office: [null],
      company: [null],
      area: [null],
      hrCode: [null],
      email: [null],
      watt: [null],
      purchasedValue: [null],
      purchasedDate: [null],
      usageDate: [null],
      recordDate: [null],
      notes: [null],
      model: [null],
      brand: [null],
      status: [null],
      state: [null],
      serial: [null],
      vendor: [null],
      division: [null],
      groupDep: [null],
      ad: [false],
      me: [false],
      tm: [false],
      detail1: [null],
      detail2: [null],
      detail3: [null],
      detail4: [null],
      detail5: [null],
      detail6: [null],
      detail7: [null],
      detail8: [null],
      detail9: [null]
    });

    this.handoverForm = this.fb.group({
      form: ['ASS', [Validators.required]],
      userTransfer: [null, [Validators.required]],
      departmentTransfer: [null],
      officeTransfer: [null],
      companyTransfer: ['YLVN'],
      areaTransfer: ['VN'],
      divisionTransfer: [null],
      groupDepTransfer: [null],
      detail2: ['GOOD'],
      detail1: [null],
      emailTransfer: [null],
      transferSubAssets: [true],
      retrieveToITD: [false]
    });

    this.problemForm = this.fb.group({
      problemSummary: [null, [Validators.required]],
      testProcedures: [null, [Validators.required]],
      solutions: [null, [Validators.required]],
      jiraCode: [null]
    });

    this.subAssetForm = this.fb.group({
      hasMonitor: [false],
      monitorBrand: [''],
      monitorModel: [''],
      monitorSerial: [''],
      monitorStatus: ['GOOD'],
      monitorState: ['001'],
      monitorVendor: [null],
      monitorWarrantyExpiration: [1],
      monitorPurchasedValue: [0],
      monitorWatt: [''],
      monitorDetail1: [null], monitorDetail2: [null], monitorDetail3: [null], monitorDetail4: [null], monitorDetail5: [null], monitorDetail6: [null], monitorDetail7: [null], monitorDetail8: [null], monitorDetail9: [null],
      
      hasUps: [false],
      upsBrand: [''],
      upsModel: [''],
      upsSerial: [''],
      upsStatus: ['GOOD'],
      upsState: ['001'],
      upsVendor: [null],
      upsWarrantyExpiration: [1],
      upsPurchasedValue: [0],
      upsWatt: [''],
      upsDetail1: [null], upsDetail2: [null], upsDetail3: [null], upsDetail4: [null], upsDetail5: [null], upsDetail6: [null], upsDetail7: [null], upsDetail8: [null], upsDetail9: [null],
      
      hasMouse: [false],
      mouseBrand: [''],
      mouseModel: [''],
      mouseSerial: [''],
      mouseStatus: ['GOOD'],
      mouseState: ['001'],
      mouseVendor: [null],
      mouseWarrantyExpiration: [1],
      mousePurchasedValue: [0],
      mouseWatt: [''],
      mouseDetail1: [null], mouseDetail2: [null], mouseDetail3: [null], mouseDetail4: [null], mouseDetail5: [null], mouseDetail6: [null], mouseDetail7: [null], mouseDetail8: [null], mouseDetail9: [null],
      
      hasKeyboard: [false],
      keyboardBrand: [''],
      keyboardModel: [''],
      keyboardSerial: [''],
      keyboardStatus: ['GOOD'],
      keyboardState: ['001'],
      keyboardVendor: [null],
      keyboardWarrantyExpiration: [1],
      keyboardPurchasedValue: [0],
      keyboardWatt: [''],
      keyboardDetail1: [null], keyboardDetail2: [null], keyboardDetail3: [null], keyboardDetail4: [null], keyboardDetail5: [null], keyboardDetail6: [null], keyboardDetail7: [null], keyboardDetail8: [null], keyboardDetail9: [null]
    });

    this.specForm = this.fb.group({
      model: [null],
      serial: [null],
      status: [null],
      state: [null],
      brand: [null],
      vendor: [null],
      warrantyExpiration: [null],
      purchasedValue: [null],
      watt: [null],
      detail1: [null],
      detail2: [null],
      detail3: [null],
      detail4: [null],
      detail5: [null],
      detail6: [null],
      detail7: [null],
      detail8: [null],
      detail9: [null]
    });

    this.filterForm = this.fb.group({
      type: [null],
      unitCode: [null],
      userName: [null],
      department: [null],
      office: [null],
      company: [null],
      area: [null],
      hrCode: [null],
      email: [null],
      status: [null],
      state: [null],
      vendor: [null],
      brand: [null],
      model: [null],
      serial: [null],
      division: [null],
      groupDep: [null],
      watt: [null],
      purchasedValue: [null]
    });

    this.addForm = this.fb.group({
      type: ['LTP', [Validators.required]],
      unitCode: [null],
      department: [null, [Validators.required]],
      userName: [null],
      office: [null, [Validators.required]],
      company: [null, [Validators.required]],
      hrCode: [null],
      email: [null],
      division: [null, [Validators.required]],
      groupDep: [null, [Validators.required]],
      area: [null, [Validators.required]],
      notes: [null],
      purchasedDate: [null],
      usageDate: [null],
      recordDate: [null],
      model: [null],
      serial: [null, [Validators.required]],
      brand: [null],
      vendor: [null],
      state: ['001'],
      status: ['GOOD'],
      warrantyExpiration: [1],
      purchasedValue: [0],
      depreciation: [0],
      value: [0],
      accessories: [false],
      watt: [null],
      detail1: [null],
      detail2: [null],
      detail3: [null],
      detail4: [null],
      detail5: [null],
      detail6: [null],
      detail7: [null],
      detail8: [null],
      detail9: [null],
      hasMonitor: [false],
      monitorBrand: [''], monitorModel: [''], monitorSerial: [''], monitorStatus: ['GOOD'], monitorState: ['001'], monitorVendor: [null], monitorWarrantyExpiration: [1], monitorPurchasedValue: [0], monitorWatt: [''],
      monitorDetail1: [null], monitorDetail2: [null], monitorDetail3: [null], monitorDetail4: [null], monitorDetail5: [null], monitorDetail6: [null], monitorDetail7: [null], monitorDetail8: [null], monitorDetail9: [null],
      hasUps: [false],
      upsBrand: [''], upsModel: [''], upsSerial: [''], upsStatus: ['GOOD'], upsState: ['001'], upsVendor: [null], upsWarrantyExpiration: [1], upsPurchasedValue: [0], upsWatt: [''],
      upsDetail1: [null], upsDetail2: [null], upsDetail3: [null], upsDetail4: [null], upsDetail5: [null], upsDetail6: [null], upsDetail7: [null], upsDetail8: [null], upsDetail9: [null],
      hasMouse: [false],
      mouseBrand: [''], mouseModel: [''], mouseSerial: [''], mouseStatus: ['GOOD'], mouseState: ['001'], mouseVendor: [null], mouseWarrantyExpiration: [1], mousePurchasedValue: [0], mouseWatt: [''],
      mouseDetail1: [null], mouseDetail2: [null], mouseDetail3: [null], mouseDetail4: [null], mouseDetail5: [null], mouseDetail6: [null], mouseDetail7: [null], mouseDetail8: [null], mouseDetail9: [null],
      hasKeyboard: [false],
      keyboardBrand: [''], keyboardModel: [''], keyboardSerial: [''], keyboardStatus: ['GOOD'], keyboardState: ['001'], keyboardVendor: [null], keyboardWarrantyExpiration: [1], keyboardPurchasedValue: [0], keyboardWatt: [''],
      keyboardDetail1: [null], keyboardDetail2: [null], keyboardDetail3: [null], keyboardDetail4: [null], keyboardDetail5: [null], keyboardDetail6: [null], keyboardDetail7: [null], keyboardDetail8: [null], keyboardDetail9: [null]
    });

    this.addForm.get('type')?.valueChanges.subscribe(val => {
      if (val) {
        this.addForm.patchValue({
          hasMonitor: false,
          hasUps: false,
          hasMouse: false,
          hasKeyboard: false
        });
        this.loadAddFormDeviceConfig(val);
        this.updateSubAssetLists();
      } else {
        this.addActiveDeviceConfig = null;
      }
    });

    this.assetForm.get('type')?.valueChanges.subscribe(val => {
      if (val && !this.selectedAsset) {
        this.subAssetForm.patchValue({
          hasMonitor: false,
          hasUps: false,
          hasMouse: false,
          hasKeyboard: false
        });
        this.autoGenerateUnitCode();
      }
    });

    // Listeners for date and company changes to auto generate Unit Code
    this.assetForm.get('purchasedDate')?.valueChanges.subscribe(() => this.autoGenerateUnitCode());
    this.assetForm.get('company')?.valueChanges.subscribe(() => this.autoGenerateUnitCode());
    this.setupHierarchyAutoUpdate();
  }

  setupHierarchyAutoUpdate() {
    this.autoUpdateHierarchy(this.assetForm);
    this.autoUpdateHierarchy(this.addForm);
    this.autoUpdateHierarchy(this.handoverForm, true);
  }

  autoUpdateHierarchy(form: FormGroup, isTransferForm = false) {
    const deptCtrl = form.get(isTransferForm ? 'departmentTransfer' : 'department');
    const divCtrl = form.get(isTransferForm ? 'divisionTransfer' : 'division');
    const groupCtrl = form.get(isTransferForm ? 'groupDepTransfer' : 'groupDep');

    if (groupCtrl) {
      groupCtrl.valueChanges.subscribe(groupDepId => {
        if (!groupDepId) return;
        const currentDivId = divCtrl?.value;
        if (currentDivId) {
          const div = this.orgRawList.find(x => x.id === currentDivId);
          if (!div || div.parentId !== groupDepId) {
            divCtrl?.setValue(null, { emitEvent: true });
          }
        }
      });
    }

    if (divCtrl) {
      divCtrl.valueChanges.subscribe(divId => {
        if (!divId) {
          if (deptCtrl && deptCtrl.value !== '00000') {
            deptCtrl.setValue('00000', { emitEvent: false });
          }
          return;
        }

        const div = this.orgRawList.find(x => x.id === divId);
        if (div && div.parentId) {
          const parentGroup = this.orgRawList.find(x => x.id === div.parentId && x.groupDivision === true);
          if (parentGroup && groupCtrl && groupCtrl.value !== parentGroup.id) {
            groupCtrl.setValue(parentGroup.id, { emitEvent: false });
          }
        }

        if (deptCtrl) {
          const currentDeptId = deptCtrl.value;
          if (currentDeptId !== '00000') {
            const dept = this.orgRawList.find(x => x.id === currentDeptId);
            if (!dept || dept.parentId !== divId) {
              deptCtrl.setValue('00000', { emitEvent: false });
            }
          } else {
            deptCtrl.setValue('00000', { emitEvent: false });
          }
        }
      });
    }

    if (deptCtrl) {
      deptCtrl.valueChanges.subscribe(deptId => {
        if (!deptId || deptId === '00000') return;
        const dept = this.orgRawList.find(x => x.id === deptId);
        if (dept && dept.parentId) {
          const parentObj = this.orgRawList.find(x => x.id === dept.parentId);
          if (parentObj) {
            if (parentObj.division === true) {
              if (divCtrl && divCtrl.value !== parentObj.id) {
                divCtrl.setValue(parentObj.id, { emitEvent: false });
              }
              if (parentObj.parentId) {
                const gpObj = this.orgRawList.find(x => x.id === parentObj.parentId);
                if (gpObj && gpObj.groupDivision === true && groupCtrl && groupCtrl.value !== gpObj.id) {
                  groupCtrl.setValue(gpObj.id, { emitEvent: false });
                }
              }
            } else if (parentObj.groupDivision === true) {
              if (groupCtrl && groupCtrl.value !== parentObj.id) {
                groupCtrl.setValue(parentObj.id, { emitEvent: false });
              }
              if (divCtrl && divCtrl.value !== null) {
                divCtrl.setValue(null, { emitEvent: false });
              }
            }
          }
        }
      });
    }
  }

  getFilteredDivisions(formName: 'add' | 'asset' | 'handover'): any[] {
    let form: FormGroup | null = null;
    let groupDepKey = 'groupDep';
    if (formName === 'add') form = this.addForm;
    else if (formName === 'asset') form = this.assetForm;
    else if (formName === 'handover') {
      form = this.handoverForm;
      groupDepKey = 'groupDepTransfer';
    }
    
    if (!form) return this.divisionList;
    const groupDepId = form.get(groupDepKey)?.value;
    if (!groupDepId) return this.divisionList;
    return this.divisionList.filter(d => d.parentId === groupDepId);
  }

  getFilteredDepartments(formName: 'add' | 'asset' | 'handover'): any[] {
    let form: FormGroup | null = null;
    let divisionKey = 'division';
    if (formName === 'add') form = this.addForm;
    else if (formName === 'asset') form = this.assetForm;
    else if (formName === 'handover') {
      form = this.handoverForm;
      divisionKey = 'divisionTransfer';
    }
    
    if (!form) return this.departmentList;
    const divisionId = form.get(divisionKey)?.value;
    if (!divisionId) return this.departmentList;
    return this.departmentList.filter(d => d.parentId === divisionId || d.id === '00000');
  }

  autoGenerateUnitCode() {
    const type = this.assetForm.get('type')?.value;
    const purchasedDate = this.assetForm.get('purchasedDate')?.value;
    const company = this.assetForm.get('company')?.value || 'YLVN';
    
    if (type && purchasedDate && !this.selectedAsset) {
      // Find area or default to South
      const area = 'South'; 
      this.assetService.generateUnitCode(type, purchasedDate, company, area).subscribe((res: any) => {
        if (res.status && res.data) {
          this.assetForm.patchValue({ unitCode: res.data }, { emitEvent: false });
        }
      });
    }
  }

  loadAssets(skipCount: boolean = false) {
    this.loading = true;
    this.cdr.markForCheck();

    const filterPayload = JSON.parse(JSON.stringify(this.currentFilter));
    filterPayload.skipCount = skipCount;
    
    if (!this.hasPermission('AllowSeeAllFromAllowedAreas')) {
      const user = this.getCurrentUser();
      const userId = user ? (user.userId || user.userID) : '';
      if (userId) {
        if (!filterPayload.advancedFilters) {
          filterPayload.advancedFilters = [];
        }
        filterPayload.advancedFilters.push({
          field: 'HRCode',
          operator: '==',
          value: userId
        });
      }
    }

    this.assetService.searchGrid(filterPayload).subscribe({
      next: (response: any) => {
        if (response.status && response.data) {
           const sorted = [...response.data].sort((a: any, b: any) => {
             return (b.unitCode || '').localeCompare(a.unitCode || '');
           });
           this.rowData = sorted;
           if (!skipCount) {
             this.totalRecords = response.totalRecords !== undefined && response.totalRecords !== null ? response.totalRecords : response.data.length;
           }
        } else {
           this.rowData = [];
           this.totalRecords = 0;
        }
        
        this.loading = false;
        this.cdr.markForCheck();
      },
      error: (err: any) => {
        console.error('Failed to load assets', err);
        
        this.rowData = [];
        this.totalRecords = 0;
        this.loading = false;
        this.cdr.markForCheck(); 
      }
    });
  }

  refreshSingleAssetRow(unitCode: string) {
    this.assetService.searchGrid({
      pageIndex: 1,
      pageSize: 1,
      advancedFilters: [{ field: 'UnitCode', operator: '==', value: unitCode }]
    }).subscribe({
      next: (res: any) => {
        if (res.status && res.data && res.data.length > 0) {
          const updatedAsset = res.data[0];
          const idx = this.rowData.findIndex(x => x.unitCode === unitCode);
          if (idx !== -1) {
            this.rowData[idx] = updatedAsset;
            this.rowData = [...this.rowData];
            this.cdr.markForCheck();
          }
        }
      }
    });
  }

  onSelectAsset(data: Asset) {
    this.router.navigate(['/assets', data.unitCode]);
  }



  showAssetDetails(data: Asset) {
    this.selectedAsset = data;
    this.selectedHandover = null;
    this.activeSpecTarget = 'main';
    this.isEditingDetails = false;
    this.activeDetailTabIndex = 0;

    const assetArea = (data as any).area;
    if (assetArea && !this.allowedAreas.includes(assetArea)) {
      this.allowedAreas = [...this.allowedAreas, assetArea];
    }
    this.assetForm.patchValue({
      ...this.selectedAsset,
      purchasedDate: this.selectedAsset.purchasedDate ? new Date(this.selectedAsset.purchasedDate) : null,
      usageDate: this.selectedAsset.usageDate ? new Date(this.selectedAsset.usageDate) : null,
      recordDate: this.selectedAsset.recordDate ? new Date(this.selectedAsset.recordDate) : null
    });
    this.assetForm.disable();
    this.specForm.disable();
    this.loadRelatedData(this.selectedAsset.unitCode);
    
    this.subAssetForm.reset();
    this.assetService.search({
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'MainAsset', operator: '==', value: data.unitCode }]
    }).subscribe((res: any) => {
      if (res.status && res.data) {
        const subAssets: any[] = (res.data || []).filter((sub: any) => sub.unitCode !== data.unitCode && (sub.mainAsset === data.unitCode || sub.MainAsset === data.unitCode));
        subAssets.forEach(sub => {
          const typeUpper = sub.type?.toUpperCase();
          const prefix = this.getSubAssetControlPrefix(typeUpper);
          const flagName = `has${prefix.charAt(0).toUpperCase() + prefix.slice(1)}`;

          if (!this.extraSubAssetTypes.includes(typeUpper) && !['MNT', 'UPS', 'MOU', 'KBO'].includes(typeUpper)) {
            this.extraSubAssetTypes.push(typeUpper);
            this.subAssetForm.addControl(flagName, this.fb.control(true));
            this.subAssetForm.addControl(`${prefix}Brand`, this.fb.control(''));
            this.subAssetForm.addControl(`${prefix}Model`, this.fb.control(''));
            this.subAssetForm.addControl(`${prefix}Serial`, this.fb.control(''));
            this.subAssetForm.addControl(`${prefix}Status`, this.fb.control('GOOD'));
            this.subAssetForm.addControl(`${prefix}State`, this.fb.control('001'));
            this.subAssetForm.addControl(`${prefix}Vendor`, this.fb.control(null));
            this.subAssetForm.addControl(`${prefix}WarrantyExpiration`, this.fb.control(1));
            this.subAssetForm.addControl(`${prefix}PurchasedValue`, this.fb.control(0));
            this.subAssetForm.addControl(`${prefix}Watt`, this.fb.control(''));
            for (let i = 1; i <= 9; i++) {
              this.subAssetForm.addControl(`${prefix}Detail${i}`, this.fb.control(null));
            }
          }

          const patchObj: any = {};
          patchObj[flagName] = true;
          patchObj[`${prefix}Brand`] = sub.brand;
          patchObj[`${prefix}Model`] = sub.model;
          patchObj[`${prefix}Serial`] = sub.serial;
          patchObj[`${prefix}Status`] = sub.status;
          patchObj[`${prefix}State`] = sub.state;
          patchObj[`${prefix}Vendor`] = sub.vendor;
          patchObj[`${prefix}WarrantyExpiration`] = sub.warrantyExpiration;
          patchObj[`${prefix}PurchasedValue`] = sub.purchasedValue;
          patchObj[`${prefix}Watt`] = sub.watt;
          for (let i = 1; i <= 9; i++) {
            patchObj[`${prefix}Detail${i}`] = sub[`detail${i}`];
          }
          this.subAssetForm.patchValue(patchObj);
        });
        this.updateSubAssetLists();
      }
      
      // Initialize specifications form with Main target values
      this.onSpecTargetChange('main');
    });
  }

  saveActiveSpecValues() {
    const vals = this.specForm.value;
    if (this.activeSpecTarget === 'main') {
      this.assetForm.patchValue({
        model: vals.model,
        serial: vals.serial,
        status: vals.status,
        state: vals.state,
        brand: vals.brand,
        vendor: vals.vendor,
        warrantyExpiration: vals.warrantyExpiration,
        purchasedValue: vals.purchasedValue,
        watt: vals.watt,
        detail1: vals.detail1,
        detail2: vals.detail2,
        detail3: vals.detail3,
        detail4: vals.detail4,
        detail5: vals.detail5,
        detail6: vals.detail6,
        detail7: vals.detail7,
        detail8: vals.detail8,
        detail9: vals.detail9
      });
    } else {
      const prefix = this.activeSpecTarget;
      const updateObj: any = {};
      updateObj[`${prefix}Model`] = vals.model;
      updateObj[`${prefix}Serial`] = vals.serial;
      updateObj[`${prefix}Status`] = vals.status;
      updateObj[`${prefix}State`] = vals.state;
      updateObj[`${prefix}Brand`] = vals.brand;
      updateObj[`${prefix}Vendor`] = vals.vendor;
      updateObj[`${prefix}WarrantyExpiration`] = vals.warrantyExpiration;
      updateObj[`${prefix}PurchasedValue`] = vals.purchasedValue;
      updateObj[`${prefix}Watt`] = vals.watt;
      updateObj[`${prefix}Detail1`] = vals.detail1;
      updateObj[`${prefix}Detail2`] = vals.detail2;
      updateObj[`${prefix}Detail3`] = vals.detail3;
      updateObj[`${prefix}Detail4`] = vals.detail4;
      updateObj[`${prefix}Detail5`] = vals.detail5;
      updateObj[`${prefix}Detail6`] = vals.detail6;
      updateObj[`${prefix}Detail7`] = vals.detail7;
      updateObj[`${prefix}Detail8`] = vals.detail8;
      updateObj[`${prefix}Detail9`] = vals.detail9;
      this.subAssetForm.patchValue(updateObj);
    }
  }

  onSpecTargetChange(target: string) {
    if (this.activeSpecTarget !== target) {
      this.saveActiveSpecValues();
    }
    this.activeSpecTarget = target;
    
    let type = '';
    if (target === 'main') {
      type = (this.selectedAsset?.type || this.assetForm.get('type')?.value || '').trim();
    } else {
      type = this.getSubAssetCodeFromTarget(target);
    }

    if (!type) {
      this.activeDeviceConfig = null;
      return;
    }

    this.typeService.getDeviceConfig(type).subscribe((res: any) => {
      if (res.status && res.data) {
        this.activeDeviceConfig = res.data;
        for (let i = 1; i <= 9; i++) {
          const groupName = this.activeDeviceConfig[`detail${i}`];
          if (groupName && groupName !== 'NONE') {
            this.techSpecService.getItemsByGroup(groupName).subscribe((itemRes: any) => {
              this.detailOptions[`detail${i}`] = itemRes.status && itemRes.data ? itemRes.data : [];
            });
          } else {
            this.detailOptions[`detail${i}`] = [];
          }
        }
      } else {
        this.activeDeviceConfig = null;
      }
    });

    if (target === 'main') {
      this.specForm.patchValue({
        model: this.assetForm.get('model')?.value,
        serial: this.assetForm.get('serial')?.value,
        status: this.assetForm.get('status')?.value,
        state: this.assetForm.get('state')?.value,
        brand: this.assetForm.get('brand')?.value,
        vendor: this.assetForm.get('vendor')?.value,
        warrantyExpiration: this.assetForm.get('warrantyExpiration')?.value,
        purchasedValue: this.assetForm.get('purchasedValue')?.value,
        watt: this.assetForm.get('watt')?.value,
        detail1: this.assetForm.get('detail1')?.value,
        detail2: this.assetForm.get('detail2')?.value,
        detail3: this.assetForm.get('detail3')?.value,
        detail4: this.assetForm.get('detail4')?.value,
        detail5: this.assetForm.get('detail5')?.value,
        detail6: this.assetForm.get('detail6')?.value,
        detail7: this.assetForm.get('detail7')?.value,
        detail8: this.assetForm.get('detail8')?.value,
        detail9: this.assetForm.get('detail9')?.value
      }, { emitEvent: false });
    } else {
      const prefix = target;
      this.specForm.patchValue({
        model: this.subAssetForm.get(`${prefix}Model`)?.value,
        serial: this.subAssetForm.get(`${prefix}Serial`)?.value,
        status: this.subAssetForm.get(`${prefix}Status`)?.value || 'GOOD',
        state: this.subAssetForm.get(`${prefix}State`)?.value || '001',
        brand: this.subAssetForm.get(`${prefix}Brand`)?.value,
        vendor: this.subAssetForm.get(`${prefix}Vendor`)?.value || this.assetForm.get('vendor')?.value,
        warrantyExpiration: this.subAssetForm.get(`${prefix}WarrantyExpiration`)?.value || 1,
        purchasedValue: this.subAssetForm.get(`${prefix}PurchasedValue`)?.value || 0,
        watt: this.subAssetForm.get(`${prefix}Watt`)?.value || '',
        detail1: this.subAssetForm.get(`${prefix}Detail1`)?.value,
        detail2: this.subAssetForm.get(`${prefix}Detail2`)?.value,
        detail3: this.subAssetForm.get(`${prefix}Detail3`)?.value,
        detail4: this.subAssetForm.get(`${prefix}Detail4`)?.value,
        detail5: this.subAssetForm.get(`${prefix}Detail5`)?.value,
        detail6: this.subAssetForm.get(`${prefix}Detail6`)?.value,
        detail7: this.subAssetForm.get(`${prefix}Detail7`)?.value,
        detail8: this.subAssetForm.get(`${prefix}Detail8`)?.value,
        detail9: this.subAssetForm.get(`${prefix}Detail9`)?.value
      }, { emitEvent: false });
    }
  }

  loadRelatedData(unitCode: string, fromDate?: Date | null, toDate?: Date | null) {
    const filter: any = {
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'unitCode', operator: '==', value: unitCode }]
    };

    this.handoverService.search(filter).subscribe((res: any) => {
      let data = res.status && res.data ? res.data : [];
      data.sort((a: any, b: any) => {
        const timeA = a.handoverDate ? new Date(a.handoverDate).getTime() : 0;
        const timeB = b.handoverDate ? new Date(b.handoverDate).getTime() : 0;
        if (timeA !== timeB) return timeB - timeA;
        return (b.id || '').localeCompare(a.id || '');
      });
      this.handoverRowData = data;
    });

    this.problemReportService.search(filter).subscribe((res: any) => {
      let data = res.status && res.data ? res.data : [];
      data.sort((a: any, b: any) => {
        const timeA = a.reportDate ? new Date(a.reportDate).getTime() : 0;
        const timeB = b.reportDate ? new Date(b.reportDate).getTime() : 0;
        if (timeA !== timeB) return timeB - timeA;
        return (b.id || '').localeCompare(a.id || '');
      });
      this.problemRowData = data;
    });

    this.loadLogs(unitCode, fromDate, toDate);
  }

  loadLogs(unitCode: string, fromDate?: Date | null, toDate?: Date | null) {
    const filter: any = {
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'unitCode', operator: '==', value: unitCode }]
    };

    if (fromDate) {
      filter.advancedFilters.push({ field: 'dateModified', operator: '>=', value: fromDate.toISOString() });
    }
    if (toDate) {
      const endOfDay = new Date(toDate);
      endOfDay.setHours(23, 59, 59, 999);
      filter.advancedFilters.push({ field: 'dateModified', operator: '<=', value: endOfDay.toISOString() });
    }

    this.assetsLogService.search(filter).subscribe((res: any) => {
      let data = res.status && res.data ? res.data : [];
      data.sort((a: any, b: any) => {
        const timeA = a.dateModified ? new Date(a.dateModified).getTime() : 0;
        const timeB = b.dateModified ? new Date(b.dateModified).getTime() : 0;
        return timeB - timeA;
      });
      this.logRowData = data;
      this.cdr.markForCheck();
    });
  }

  onFilter() {
    this.isFilterVisible = true;
  }

  initializeDefaultFilters() {
    if (this.isViewer()) {
      this.selectedFilters = [];
    } else {
      this.selectedFilters = [
        {
          id: 1,
          field: 'Type',
          fieldName: 'Type',
          dataType: 'multiselect',
          multiSelectValue: ['DPC', 'LTP', 'dpc', 'ltp'],
          selectDatas: []
        },
        {
          id: 2,
          field: 'State',
          fieldName: 'State',
          dataType: 'multiselect',
          multiSelectValue: ['001', '002', '003'],
          selectDatas: []
        }
      ];
    }
    this.initializeFilterFields();
  }

  initializeFilterFields() {
    this.filterFields = [
      { field: 'Type', fieldName: 'Type', dataType: 'multiselect', selectDatas: this.typeList.map(t => ({ id: t.type, text: `${t.typeName} (${t.type})` })) },
      { field: 'UnitCode', fieldName: 'Unit Code', dataType: 'string' },
      { field: 'UserName', fieldName: 'User', dataType: 'string' },
      { field: 'Department', fieldName: 'Department', dataType: 'multiselect', selectDatas: this.filterDepartmentList.map(d => ({ id: d.id, text: d.name })) },
      { field: 'Office', fieldName: 'Office', dataType: 'multiselect', selectDatas: this.officeList.map(o => ({ id: o.id, text: o.name })) },
      { field: 'Company', fieldName: 'Company', dataType: 'multiselect', selectDatas: this.companyList.map(c => ({ id: c.id, text: c.name })) },
      { field: 'Area', fieldName: 'Area', dataType: 'multiselect', selectDatas: this.allowedAreas.map(a => ({ id: a, text: a })) },
      { field: 'HRCode', fieldName: 'HR Code', dataType: 'string' },
      { field: 'Email', fieldName: 'Email', dataType: 'string' },
      { field: 'Status', fieldName: 'Status', dataType: 'multiselect', selectDatas: [
        { id: 'GOOD', text: 'GOOD' },
        { id: 'NORMAL', text: 'NORMAL' },
        { id: 'BAD', text: 'BAD' },
        { id: 'BROKEN', text: 'BROKEN' }
      ]},
      { field: 'State', fieldName: 'State', dataType: 'multiselect', selectDatas: this.stateList.map(s => ({ id: s.id, text: s.description || s.state })) },
      { field: 'Vendor', fieldName: 'Vendor', dataType: 'multiselect', selectDatas: this.vendorList.map(v => ({ id: v.id, text: v.name })) },
      { field: 'Brand', fieldName: 'Brand', dataType: 'string' },
      { field: 'Model', fieldName: 'Model', dataType: 'string' },
      { field: 'Serial', fieldName: 'Serial', dataType: 'string' },
      { field: 'Division', fieldName: 'Division', dataType: 'multiselect', selectDatas: this.filterDivisionList.map(d => ({ id: d.id, text: d.name })) },
      { field: 'GroupDep', fieldName: 'Group Dep', dataType: 'multiselect', selectDatas: this.groupDepList.map(g => ({ id: g.id, text: g.name })) },
      { field: 'Watt', fieldName: 'Watt', dataType: 'number' },
      { field: 'PurchasedValue', fieldName: 'Purchased Value', dataType: 'number' }
    ];

    this.selectedFilters.forEach(item => {
      const fieldConfig = this.filterFields.find(f => f.field === item.field);
      if (fieldConfig) {
        item.selectDatas = fieldConfig.selectDatas;
      }
    });
  }

  changeFilterField(field: string, item: any) {
    const fieldExist = this.filterFields.find(x => x.field === field);
    if (fieldExist) {
      item.dataType = fieldExist.dataType;
      item.fieldName = fieldExist.fieldName;
      item.selectDatas = fieldExist.selectDatas;
      item.stringValue = '';
      item.numberValue = null;
      item.dropdownValue = null;
      item.multiSelectValue = [];
    }
  }

  addFilterRow() {
    this.selectedFilters.push({
      id: new Date().getTime(),
      field: null,
      dataType: 'string',
      selectDatas: [],
      multiSelectValue: []
    });
  }

  deleteFilterRow(index: number) {
    this.selectedFilters.splice(index, 1);
  }

  submitAdvancedFilter() {
    const payload: any = {
      pageIndex: 1,
      pageSize: this.currentFilter.pageSize || 20,
      orderBy: 'PurchasedDate desc',
      advancedFilters: []
    };

    this.selectedFilters.forEach(item => {
      if (!item.field) return;

      let value: string = '';
      let operator: string = '==';

      switch (item.dataType) {
        case 'string':
          value = item.stringValue || '';
          operator = 'contains';
          break;
        case 'number':
          value = item.numberValue !== null && item.numberValue !== undefined ? item.numberValue.toString() : '';
          operator = '==';
          break;
        case 'dropdown':
          value = item.dropdownValue || '';
          operator = '==';
          break;
        case 'multiselect':
          if (item.multiSelectValue && item.multiSelectValue.length > 0) {
            value = item.multiSelectValue.join('/');
            operator = '==';
          }
          break;
      }

      if (value) {
        payload.advancedFilters.push({
          field: item.field,
          operator: operator,
          value: value
        });
      }
    });

    this.currentFilter = payload;
    this.loadAssets();
    this.isFilterVisible = false;
  }

  resetAdvancedFilter() {
    this.initializeDefaultFilters();
  }

  @ViewChild('disposalFileInput') disposalFileInput!: ElementRef;

  onDisposal() {
    if (!this.hasPermission('Disposal')) {
      this.message.error('You do not have permission to access disposal.');
      return;
    }
    this.isDisposalModalVisible = true;
    if (this.selectedAssetCodes && this.selectedAssetCodes.size > 0) {
      this.disposalInputCodes = Array.from(this.selectedAssetCodes).join('\n');
    } else {
      this.disposalInputCodes = '';
    }
    this.disposalCheckedAssets = [];
    this.disposalWfdNote = '';
    this.disposalConfirmNote = '';
    this.selectedDisposalBatch = null;
    this.disposalBatchDetails = [];
    this.isDisposalTransferEnabled = false;
    this.loadDisposalBatches();
  }

  triggerDisposalUpload() {
    this.disposalFileInput.nativeElement.click();
  }

  onDisposalFileSelected(event: any) {
    const file: File = event.target.files[0];
    if (file) {
      this.disposalService.importUnitCodes(file).subscribe({
        next: (res: any) => {
          if (res.status && res.data) {
            this.disposalInputCodes = res.data.join('\n');
            this.checkDisposalUnitCodes();
            this.message.success('Đã import mã unit code từ Excel thành công!');
          } else {
            this.message.error(res.message || 'Lỗi khi import file Excel!');
          }
          // Reset file input value to allow selecting same file again
          event.target.value = '';
        },
        error: (err: any) => {
          this.message.error('Lỗi khi import file Excel!');
          event.target.value = '';
        }
      });
    }
  }

  loadDisposalBatches() {
    this.disposalService.getBatches().subscribe({
      next: (res: any) => {
        this.disposalBatches = res.status && res.data ? res.data : [];
        if (this.disposalBatches.length > 0 && !this.selectedDisposalBatch) {
          this.selectDisposalBatch(this.disposalBatches[0]);
        }
      }
    });
  }

  checkDisposalUnitCodes() {
    if (!this.disposalInputCodes) {
      this.message.warning('Vui lòng nhập danh sách Unit Code!');
      return;
    }
    const codes = this.disposalInputCodes
      .split(/[\n,\t]+/)
      .map(c => c.trim())
      .filter(c => c.length > 0);

    if (codes.length === 0) {
      this.message.warning('Danh sách Unit Code trống!');
      return;
    }

    this.disposalService.checkUnitCodes(codes).subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.disposalCheckedAssets = res.data;
          const allValid = this.disposalCheckedAssets.every(a => a.valid);
          this.isDisposalTransferEnabled = allValid && this.disposalCheckedAssets.length > 0;
        } else {
          this.message.error('Không thể kiểm tra danh sách!');
        }
      }
    });
  }

  transferDisposalWfd() {
    if (!this.isDisposalTransferEnabled || this.disposalCheckedAssets.length === 0) {
      return;
    }
    const unitCodes = this.disposalCheckedAssets.map(a => a.unitCode);
    const currentUser = this.getCurrentUser();
    const creator = currentUser ? currentUser.userId : 'admin';
    this.disposalService.transferWfd(unitCodes, this.disposalWfdNote, creator).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success('Chuyển sang trạng thái Chờ thanh lý thành công!');
          this.disposalInputCodes = '';
          this.disposalCheckedAssets = [];
          this.disposalWfdNote = '';
          this.isDisposalTransferEnabled = false;
          this.loadDisposalBatches();
          this.loadAssets();
        } else {
          this.message.error(res.message || 'Lỗi khi chuyển trạng thái!');
        }
      }
    });
  }

  selectDisposalBatch(batch: any) {
    this.selectedDisposalBatch = batch;
    this.disposalService.getBatchDetails(batch.idDisposal).subscribe({
      next: (res: any) => {
        this.disposalBatchDetails = res.status && res.data ? res.data : [];
      }
    });
  }

  confirmDisposal() {
    if (!this.selectedDisposalBatch) return;
    const currentUser = this.getCurrentUser();
    const creator = currentUser ? currentUser.userId : 'admin';
    this.disposalService.confirmDispose(this.selectedDisposalBatch.idDisposal, this.disposalConfirmNote, creator).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success('Xác nhận thanh lý thành công!');
          this.disposalConfirmNote = '';
          this.loadDisposalBatches();
          this.loadAssets();
        } else {
          this.message.error(res.message || 'Lỗi khi xác nhận thanh lý!');
        }
      }
    });
  }

  onImport() {
    if (!this.hasPermission('AssetsModify')) {
      this.message.error('You do not have permission to modify assets.');
      return;
    }
    this.fileInput.nativeElement.click();
  }

  onFileSelected(event: any) {
    const file: File = event.target.files[0];
    if (file) {
      this.assetService.importFile(file).subscribe({
        next: (res: any) => {
          if (res.status) {
            if (res.data) {
              // Failed rows present, download error log file
              const byteCharacters = atob(res.data);
              const byteNumbers = new Array(byteCharacters.length);
              for (let i = 0; i < byteCharacters.length; i++) {
                byteNumbers[i] = byteCharacters.charCodeAt(i);
              }
              const byteArray = new Uint8Array(byteNumbers);
              const blob = new Blob([byteArray], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
              saveAs(blob, 'Import_Errors.xlsx');
              this.message.warning(res.message || 'Some rows failed validation. Error report downloaded.');
            } else {
              this.message.success(res.message || 'File imported successfully');
            }
            this.loadAssets();
          } else {
            this.message.error(res.message || 'Import failed');
          }
        },
        error: (err: any) => {
          console.error('Import error', err);
          this.message.error('An error occurred during import');
        }
      });
      
      this.fileInput.nativeElement.value = '';
    }
  }

  onDownloadTemplate() {
    this.assetService.downloadTemplate().subscribe({
      next: (blob: any) => {
        saveAs(blob, 'AssetsCatalouges.xlsx');
      },
      error: (err: any) => {
        console.error('Download template error', err);
        this.message.error('Failed to download template file');
      }
    });
  }

  selectedAssetCodes = new Set<string>();

  isAllChecked(): boolean {
    if (!this.filteredRowData || this.filteredRowData.length === 0) return false;
    return this.filteredRowData.every(item => this.selectedAssetCodes.has(item.unitCode));
  }

  isIndeterminate(): boolean {
    if (!this.filteredRowData || this.filteredRowData.length === 0) return false;
    const hasChecked = this.filteredRowData.some(item => this.selectedAssetCodes.has(item.unitCode));
    const allChecked = this.isAllChecked();
    return hasChecked && !allChecked;
  }

  onAllChecked(checked: boolean): void {
    if (!this.filteredRowData) return;
    this.filteredRowData.forEach(item => {
      if (checked) {
        this.selectedAssetCodes.add(item.unitCode);
      } else {
        this.selectedAssetCodes.delete(item.unitCode);
      }
    });
  }

  onItemChecked(unitCode: string, checked: boolean): void {
    if (checked) {
      this.selectedAssetCodes.add(unitCode);
    } else {
      this.selectedAssetCodes.delete(unitCode);
    }
  }

  onPrintSelectedUC() {
    const selectedAssets = this.rowData.filter(a => this.selectedAssetCodes.has(a.unitCode));
    if (selectedAssets.length === 0) {
      this.message.warning('No assets selected to print.');
      return;
    }

    const printWindow = window.open('', '', 'height=600,width=800');
    if (printWindow) {
      printWindow.document.write('<html><head><title>Print selected labels</title>');
      
      const styles = `
        <style>
          @page { 
            size: 5cm 3cm; 
            margin: 0; 
          }
          body { 
            margin: 0; 
            padding: 0; 
          }
          .print-card { 
            width: 5cm; 
            height: 3cm; 
            box-sizing: border-box; 
            padding: 0.4cm; 
            position: relative; 
            font-family: Arial, sans-serif; 
            background: #ffffff;
            border: none; /* Bỏ viền dashed khi in thật để tránh lem */
            /* Quan trọng: Ngắt trang sau mỗi tem để khớp với gap của cuộn giấy */
            page-break-after: always; 
            break-after: page; 
          }
          /* ... (Giữ nguyên các class .print-qr, .print-info, v.v. của bạn) ... */
          .print-qr { width:1.3cm; height:1.3cm; position:absolute; top:0.35cm; left:0.35cm; } 
          .print-info { position:absolute; top:0.35cm; left:1.8cm; right:0.35cm; font-size:12px; line-height:1.2; word-break:break-all; color:#000000; text-align:left; margin:0; padding:0; } 
          .print-unit-code { font-size:12px; font-weight:bold; margin:0 0 2px 0; padding:0; line-height:1.1; color:#000000; } 
          .print-info-item { font-size:12px; margin-bottom:1.5px; }
        </style>
      `;
      printWindow.document.write(styles);
      printWindow.document.write('</head><body>');

      selectedAssets.forEach(asset => {
        const qrDataString = `Unit ${asset.unitCode || ''}| Vendor ${asset.vendor || ''}| Serial ${asset.serial || ''}| PurchasedDate ${asset.purchasedDate ? new Date(asset.purchasedDate).toLocaleDateString('en-GB') : ''}| Status ${asset.status || ''}`;
        const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(qrDataString)}`;
        const purDateStr = asset.purchasedDate ? `Pur. Date: ${new Date(asset.purchasedDate).toLocaleDateString('en-GB')}` : 'Pur. Date: N/A';
        const vendorStr = `Vendor: ${this.getVendorName(asset.vendor) || asset.vendor || 'N/A'}`;
        const serialStr = `S/N: ${asset.serial || 'N/A'}`;

        printWindow.document.write(`
          <div class="print-card">
            <img src="${qrUrl}" class="print-qr" />
            <div class="print-info">
              <div class="print-unit-code">${asset.unitCode}</div>
              <div class="print-info-item">${serialStr}</div>
              <div class="print-info-item">${purDateStr}</div>
              <div class="print-info-item">${vendorStr}</div>
            </div>
          </div>
        `);
      });

      printWindow.document.write('</body></html>');
      printWindow.document.close();
      printWindow.focus();
      
      setTimeout(() => {
        printWindow.print();
        printWindow.close();
      }, 1000); 
    }
}

  onExportCurrent() {
    
    if (!this.rowData || this.rowData.length === 0) {
      this.message.warning('Không có dữ liệu để xuất Excel!');
      return; 
    }
    
    if (this.selectedAssetCodes.size > 0) {
      const selectedCodes = Array.from(this.selectedAssetCodes).join('/');
      const selectedFilter = {
        pageIndex: 1,
        pageSize: 1000,
        advancedFilters: [
          { field: 'UnitCode', operator: '==', value: selectedCodes }
        ]
      };
      this.exportWithFilter(selectedFilter, 'AssetsExport_Selected.xlsx');
    } else {
      this.exportWithFilter(this.currentFilter, 'AssetsExport_Current.xlsx');
    }
  }

  private exportWithFilter(filter: any, filename: string) {
     const filterPayload = JSON.parse(JSON.stringify(filter));

     this.assetService.exportExcel(filterPayload).subscribe({
       next: (blob: any) => {
         saveAs(blob, filename);
       },
       error: (err: any) => {
         console.error('Export error', err);
         this.message.error('An error occurred during export');
       }
     });
  }

  onRefresh() {
    if (this.isViewer()) {
      this.currentFilter = {
        pageIndex: 1,
        pageSize: 20,
        orderBy: 'PurchasedDate desc',
        advancedFilters: []
      };
    } else {
      this.currentFilter = {
        pageIndex: 1,
        pageSize: 20,
        orderBy: 'PurchasedDate desc',
        advancedFilters: [
          { field: 'Type', operator: '==', value: 'DPC/LTP/dpc/ltp' },
          { field: 'State', operator: '==', value: '001/002/003' }
        ]
      };
    }
    this.searchText = '';
    this.currentFilter.searchText = '';
    this.searchSubject.next('');
    this.loadAssets(false);
  }



  getCurrentUser(): any {
    const userStr = localStorage.getItem('currentUser');
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch (e) {
        return null;
      }
    }
    return null;
  }

  hasPermission(permission: string): boolean {
    return this.userService.hasPermission(permission);
  }

  isViewer(): boolean {
    const user = this.getCurrentUser();
    if (!user) return false;
    const roles: string[] = user.roles || user.Roles || [];
    const isAdmin = roles.some(r => r.toLowerCase() === 'admin');
    if (isAdmin) return false;
    const perms: string[] = user.permissions || user.Permissions || [];
    const elevatedPerms = ['AssetsModify', 'AssetsDel', 'DepartmentModify', 'DepartmentDel',
      'TypeModify', 'TypeDel', 'UserManager', 'Disposal', 'PermissionApproval',
      'AllowSeeAllFromAllowedAreas'];
    return !elevatedPerms.some(p => perms.includes(p));
  }

  onAdd() {
    if (!this.hasPermission('AssetsModify')) {
      this.message.error('You do not have permission to modify assets.');
      return;
    }
    this.isAddModalVisible = true;
    this.currentAddStep = 0;
    this.resetSubAssetStates();
    this.loadSubAssetDeviceConfigs();
    const today = new Date();
    const user = this.getCurrentUser();
    this.addForm.reset({
      type: 'LTP',
      department: user?.department || user?.Department,
      office: null,
      company: null,
      userName: 'IT',
      email: '',
      hrCode: 'ITD',
      model: null,
      brand: null,
      serial: null,
      vendor: null,
      purchasedDate: today,
      usageDate: today,
      recordDate: today,
      status: 'GOOD',
      state: '001',
      warrantyExpiration: 1,
      purchasedValue: 0,
      depreciation: 0,
      value: 0,
      accessories: false,
      detail1: null,
      detail2: null,
      detail3: null,
      detail4: null,
      detail5: null,
      detail6: null,
      detail7: null,
      detail8: null,
      detail9: null,
      area: null,
      
      hasMonitor: false,
      monitorStatus: 'GOOD', monitorState: '001', monitorWarrantyExpiration: 1, monitorPurchasedValue: 0,
      hasUps: false,
      upsStatus: 'GOOD', upsState: '001', upsWarrantyExpiration: 1, upsPurchasedValue: 0,
      hasMouse: false,
      mouseStatus: 'GOOD', mouseState: '001', mouseWarrantyExpiration: 1, mousePurchasedValue: 0,
      hasKeyboard: false,
      keyboardStatus: 'GOOD', keyboardState: '001', keyboardWarrantyExpiration: 1, keyboardPurchasedValue: 0
    });

    this.addActiveDeviceConfig = null;
  }

  scrollToStep(step: number) {
    this.currentAddStep = step;
    this.cdr.detectChanges();
    setTimeout(() => {
      if (this.stepsWrapper) {
        const element = this.stepsWrapper.nativeElement;
        let top = 0;
        if (step === 1) {
          top = 300;
        } else if (step === 2) {
          top = 300 + 550;
        }
        element.scrollTo({
          top: top,
          behavior: 'smooth'
        });
      }
    }, 100);
  }

  onStepScroll(event: any) {
    const element = event.target;
    const scrollTop = element.scrollTop;
    let newStep = 0;

    if (scrollTop < 150) {
      newStep = 0;
    } else if (scrollTop < 575) {
      newStep = 1;
    } else {
      newStep = 2;
    }

    if (newStep !== this.currentAddStep && newStep >= 0 && newStep <= 2) {
      this.currentAddStep = newStep;
      this.cdr.detectChanges();
    }
  }

  nextAddStep() {
    if (this.currentAddStep === 0) {
      const typeCtrl = this.addForm.get('type');
      const companyCtrl = this.addForm.get('company');
      const areaCtrl = this.addForm.get('area');
      
      typeCtrl?.markAsDirty();
      typeCtrl?.updateValueAndValidity();
      companyCtrl?.markAsDirty();
      companyCtrl?.updateValueAndValidity();
      areaCtrl?.markAsDirty();
      areaCtrl?.updateValueAndValidity();

      if (typeCtrl?.invalid || companyCtrl?.invalid || areaCtrl?.invalid) {
        this.message.error('Please fill in all required fields for Step 1');
        return;
      }
      this.scrollToStep(1);
    } else if (this.currentAddStep === 1) {
      const serialCtrl = this.addForm.get('serial');
      serialCtrl?.markAsDirty();
      serialCtrl?.updateValueAndValidity();

      if (serialCtrl?.invalid) {
        this.message.error('Please enter Serial Number');
        return;
      }

      const serial = this.addForm.get('serial')?.value || '';
      const model = this.addForm.get('model')?.value || '';
      const brand = this.addForm.get('brand')?.value || '';

      const checkFilter = {
        pageIndex: 1,
        pageSize: 1,
        advancedFilters: [
          { field: 'Serial', operator: '==', value: serial },
          { field: 'Model', operator: '==', value: model },
          { field: 'Brand', operator: '==', value: brand }
        ]
      };

      this.loading = true;
      this.assetService.search(checkFilter).subscribe({
        next: (res: any) => {
          this.loading = false;
          if (res.status && res.data && res.data.length > 0) {
            this.message.error('Serial number already exists for this Model and Brand!');
          } else {
            this.scrollToStep(2);
          }
          this.cdr.detectChanges();
        },
        error: (err: any) => {
          this.loading = false;
          this.message.error('Failed to verify serial number uniqueness');
          this.cdr.detectChanges();
        }
      });
    } else if (this.currentAddStep === 2) {
      // Done, no next step from step 3 via next button
    }
  }

  prevAddStep() {
    if (this.currentAddStep > 0) {
      this.scrollToStep(this.currentAddStep - 1);
    }
  }

  getDepartmentName(id: string): string {
    return this.departmentList.find(d => d.id === id)?.name || '';
  }

  getOfficeName(id: string): string {
    return this.officeList.find(o => o.id === id)?.name || '';
  }

  getCompanyName(id: string): string {
    return this.companyList.find(c => c.id === id)?.name || '';
  }

  getVendorName(id: string): string {
    return this.vendorList.find(v => v.id === id)?.name || '';
  }

  getStateName(id: string): string {
    return this.stateList.find(s => s.id === id)?.description || '';
  }

  getDivisionName(id: string): string {
    return this.divisionList.find(d => d.id === id)?.name || '';
  }

  getGroupDepName(id: string): string {
    return this.groupDepList.find(g => g.id === id)?.name || '';
  }



  getTechSpecLabel(target: string, num: number): string {
    if (target === 'main') {
      return this.addActiveDeviceConfig ? this.addActiveDeviceConfig['detail' + num] : '';
    }
    const code = this.getSubAssetCodeFromTarget(target);
    const config = this.subAssetConfigsData[code]?.config;
    return config ? config['detail' + num] : '';
  }

  getTechSpecValueLabel(target: string, num: number): string {
    const controlName = target === 'main' ? 'detail' + num : target + 'Detail' + num;
    const value = this.formVal[controlName];
    if (!value) return '';
    if (target === 'main') {
      const opt = this.addDetailOptions['detail' + num]?.find((o: any) => o.id === value);
      return opt ? opt.techSpecs : '';
    }
    const code = this.getSubAssetCodeFromTarget(target);
    const opt = this.subAssetConfigsData[code]?.detailOptions['detail' + num]?.find((o: any) => o.id === value);
    return opt ? opt.techSpecs : '';
  }

  hasAnyTechSpecs(target: string): boolean {
    for (let i = 1; i <= 9; i++) {
      if (this.getTechSpecValueLabel(target, i)) {
        return true;
      }
    }
    return false;
  }

  loadAddFormDeviceConfig(type: string) {
    if (!type) {
      this.addActiveDeviceConfig = null;
      return;
    }
    this.typeService.getDeviceConfig(type).subscribe((res: any) => {
      if (res.status && res.data) {
        this.addActiveDeviceConfig = res.data;
        for (let i = 1; i <= 9; i++) {
          const groupName = this.addActiveDeviceConfig[`detail${i}`];
          if (groupName && groupName !== 'NONE') {
            this.techSpecService.getItemsByGroup(groupName).subscribe((itemRes: any) => {
              this.addDetailOptions[`detail${i}`] = itemRes.status && itemRes.data ? itemRes.data : [];
            });
          } else {
            this.addDetailOptions[`detail${i}`] = [];
          }
        }
      } else {
        this.addActiveDeviceConfig = null;
      }
    });
  }

  onAddFormTypeChange(type: string) {
    this.loadAddFormDeviceConfig(type);
  }

  submitAddForm() {
    if (this.addForm.invalid) {
      Object.values(this.addForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      this.message.error('Vui lòng nhập đầy đủ các thông tin bắt buộc!');

      const typeCtrl = this.addForm.get('type');
      const companyCtrl = this.addForm.get('company');
      const areaCtrl = this.addForm.get('area');
      if (typeCtrl?.invalid || companyCtrl?.invalid || areaCtrl?.invalid) {
        this.scrollToStep(0);
      } else {
        const serialCtrl = this.addForm.get('serial');
        if (serialCtrl?.invalid) {
          this.scrollToStep(1);
        }
      }
      return;
    }

    this.currentAddStep = 3;
    this.cdr.detectChanges();
  }

  executeSubmitAddForm() {
    if (this.addForm.invalid) {
      Object.values(this.addForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      this.message.error('Vui lòng nhập đầy đủ các thông tin bắt buộc!');

      const typeCtrl = this.addForm.get('type');
      const companyCtrl = this.addForm.get('company');
      const areaCtrl = this.addForm.get('area');
      if (typeCtrl?.invalid || companyCtrl?.invalid || areaCtrl?.invalid) {
        this.scrollToStep(0);
      } else {
        const serialCtrl = this.addForm.get('serial');
        if (serialCtrl?.invalid) {
          this.scrollToStep(1);
        }
      }
      return;
    }

    this.isAddModalVisible = false;
    const vals = this.addForm.value;

    const payload = {
      type: vals.type,
      unitCode: vals.unitCode,
      department: vals.department || (this.getCurrentUser()?.department || this.getCurrentUser()?.Department || '00002'),
      userName: vals.userName || 'admin',
      office: vals.office,
      company: vals.company,
      hrCode: vals.hrCode,
      email: vals.email,
      division: vals.division,
      groupDep: vals.groupDep,
      area: vals.area || 'South',
      notes: vals.notes,
      purchasedDate: vals.purchasedDate,
      usageDate: vals.usageDate,
      recordDate: vals.recordDate,
      model: vals.model,
      serial: vals.serial,
      brand: vals.brand,
      vendor: vals.vendor,
      state: vals.state,
      status: vals.status,
      warrantyExpiration: vals.warrantyExpiration,
      purchasedValue: vals.purchasedValue,
      depreciation: vals.depreciation !== undefined && vals.depreciation !== null ? vals.depreciation : 0,
      value: vals.value !== undefined && vals.value !== null ? vals.value : 0,
      accessories: vals.accessories !== undefined && vals.accessories !== null ? vals.accessories : false,
      watt: vals.watt,
      detail1: vals.detail1, detail2: vals.detail2, detail3: vals.detail3,
      detail4: vals.detail4, detail5: vals.detail5, detail6: vals.detail6,
      detail7: vals.detail7, detail8: vals.detail8, detail9: vals.detail9,
      SubAssets: [] as any[]
    };

    const getResolvedTypeCode = (candidate: string) => {
      const match = this.typeList.find(t => t.type?.toUpperCase() === candidate.toUpperCase());
      if (match) return match.type;
      if (candidate.toUpperCase() === 'MNT') {
        const monMatch = this.typeList.find(t => t.type?.toUpperCase() === 'MON');
        if (monMatch) return monMatch.type;
      }
      return null;
    };

    const defaultSubConfigs = [
      { type: 'MON', prefix: 'monitor', hasFlag: !!vals.hasMonitor },
      { type: 'UPS', prefix: 'ups', hasFlag: !!vals.hasUps },
      { type: 'MOU', prefix: 'mouse', hasFlag: !!vals.hasMouse },
      { type: 'KBO', prefix: 'keyboard', hasFlag: !!vals.hasKeyboard }
    ];

    const extraSubConfigs = this.extraSubAssetTypes.map(code => {
      const prefix = this.getSubAssetControlPrefix(code);
      const flagName = `has${prefix.charAt(0).toUpperCase() + prefix.slice(1)}`;
      return { type: code, prefix: prefix, hasFlag: !!vals[flagName] };
    });

    const subAssetConfigs = [...defaultSubConfigs, ...extraSubConfigs];

    subAssetConfigs.forEach(config => {
      const { type, prefix, hasFlag } = config;
      if (!hasFlag) return;
      const actualType = getResolvedTypeCode(type);
      if (!actualType) return;

      const brand = vals[`${prefix}Brand`];
      const model = vals[`${prefix}Model`];
      const serial = vals[`${prefix}Serial`];

      const subAsset: any = {
        type: actualType,
        brand: brand,
        model: model,
        serial: serial,
        status: vals[`${prefix}Status`] || 'GOOD',
        state: vals[`${prefix}State`] || '001',
        vendor: vals[`${prefix}Vendor`] || vals.vendor,
        warrantyExpiration: vals[`${prefix}WarrantyExpiration`] || 1,
        purchasedValue: vals[`${prefix}PurchasedValue`] || 0,
        depreciation: 0,
        value: 0,
        accessories: false,
        watt: vals[`${prefix}Watt`] || '',
        
        department: vals.department,
        office: vals.office,
        company: vals.company,
        area: vals.area,
        division: vals.division,
        groupDep: vals.groupDep,
        userName: vals.userName,
        email: vals.email,
        hrCode: vals.hrCode,
        purchasedDate: vals.purchasedDate,
        usageDate: vals.usageDate,
        recordDate: vals.recordDate,
        notes: vals.notes
      };
      for (let i = 1; i <= 9; i++) {
        subAsset[`detail${i}`] = vals[`${prefix}Detail${i}`];
      }
      payload.SubAssets.push(subAsset);
    });

    this.assetService.create(payload).subscribe({
      next: (res: any) => {
        if (res.status) {
          const newUnitCode = res.data?.unitCode || payload.unitCode || '';
          this.message.success(`Asset created successfully! Unit Code: ${newUnitCode}`, { nzDuration: 3000 });
          this.loadAssets();
          this.isAddModalVisible = false;
          this.currentAddStep = 0;
          this.addForm.reset();
          this.router.navigate([], { queryParams: { id: newUnitCode }, queryParamsHandling: 'merge' });
        } else {
          this.message.error('Failed to add asset');
        }
      },
      error: (err: any) => {
        console.error(err);
        this.message.error(err.error?.message || 'An error occurred while adding asset');
      }
    });
  }

    onPrintUC() {
    if (!this.selectedAsset) {
      this.message.warning('Please select an asset to print.');
      return;
    }
    
    const qrDataString = `Unit ${this.selectedAsset.unitCode || ''}| Vendor ${this.getVendorName(this.selectedAsset.vendor) || this.selectedAsset.vendor || ''}| Serial ${this.selectedAsset.serial || ''}| PurchasedDate ${this.selectedAsset.purchasedDate ? new Date(this.selectedAsset.purchasedDate).toLocaleDateString('en-GB') : ''}| Status ${this.selectedAsset.status || ''}`;

    this.printData = {
      unitCode: this.selectedAsset.unitCode,
      purchasedDate: this.selectedAsset.purchasedDate ? `Pur. Date: ${new Date(this.selectedAsset.purchasedDate).toLocaleDateString('en-GB')}` : 'Pur. Date: N/A',
      vendor: `Vendor: ${this.getVendorName(this.selectedAsset.vendor) || this.selectedAsset.vendor || 'N/A'}`,
      serial: `S/N: ${this.selectedAsset.serial || 'N/A'}`,
      qrUrl: `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(qrDataString)}`
    };
    
    this.isPrintVisible = true;
  }

  printBarcode() {
    const printContents = document.getElementById('print-section')?.innerHTML;
    if (printContents) {
      const printWindow = window.open('', '', 'height=600,width=800');
      if (printWindow) {
        printWindow.document.write('<html><head><title>Print UC</title>');
        printWindow.document.write('<style>body{margin:0;padding:0;} .print-wrapper{padding:0 !important;background:none !important;} .print-card { width:5cm; height:3cm; box-sizing:border-box; padding:0.35cm; position:relative; font-family:Arial,sans-serif; background:#ffffff; box-shadow:none !important; border:none !important; } .print-qr { width:1.3cm; height:1.3cm; position:absolute; top:0.35cm; left:0.35cm; } .print-info { position:absolute; top:0.35cm; left:1.8cm; right:0.35cm; font-size:13px; line-height:1.25; word-break:break-all; color:#000000; text-align:left; margin:0; padding:0; } .print-unit-code { font-size:13px; font-weight:bold; margin:0 0 2px 0; padding:0; line-height:1.1; color:#000000; } .print-info-item { font-size:13px; margin-bottom:1.5px; } @page { size:5cm 3cm; margin:0; }</style>');
        printWindow.document.write('</head><body>');
        printWindow.document.write(printContents);
        printWindow.document.write('</body></html>');
        printWindow.document.close();
        printWindow.focus();
        setTimeout(() => {
          printWindow.print();
          printWindow.close();
        }, 500);
      }
    }
  }

    getTypeIcon(type: string): string {
    if (!type) return 'devices';
    const lower = type.toLowerCase();
    if (lower.includes('ltp') || lower.includes('laptop')) return 'laptop_mac';
    if (lower.includes('mnt') || lower.includes('monitor')) return 'monitor';
    if (lower.includes('srv') || lower.includes('server')) return 'dns';
    if (lower.includes('mou') || lower.includes('mouse')) return 'mouse';
    if (lower.includes('kbo') || lower.includes('keyboard')) return 'keyboard';
    if (lower.includes('ups')) return 'battery_charging_full';
    if (lower.includes('router') || lower.includes('swi')) return 'router';
    return 'devices';
  }

  getInitials(name: string): string {
    if (!name) return '--';
    const parts = name.trim().split(/\s+/);
    if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
    const first = parts[0].charAt(0);
    const last = parts[parts.length - 1].charAt(0);
    return (first + last).toUpperCase();
  }

  getStateBadgeClass(stateName: string): string {
    if (!stateName) return 'state-badge-default';
    const lower = stateName.toLowerCase();
    if (lower.includes('active') || lower.includes('assigned') || lower.includes('đang sử dụng') || lower.includes('hợp lệ')) {
      return 'state-badge-active';
    }
    if (lower.includes('service') || lower.includes('maintenance') || lower.includes('bảo trì')) {
      return 'state-badge-service';
    }
    if (lower.includes('storage') || lower.includes('kho') || lower.includes('lưu kho')) {
      return 'state-badge-storage';
    }
    if (lower.includes('retired') || lower.includes('disposed') || lower.includes('hủy') || lower.includes('thu hồi')) {
      return 'state-badge-retired';
    }
    return 'state-badge-default';
  }

  getStatusBadgeClass(statusName: string): string {
    if (!statusName) return 'status-badge-default';
    const lower = statusName.toLowerCase();
    if (lower === 'good') {
      return 'status-badge-good';
    }
    if (lower === 'normal') {
      return 'status-badge-normal';
    }
    if (lower === 'bad') {
      return 'status-badge-bad';
    }
    if (lower === 'broken') {
      return 'status-badge-broken';
    }
    return 'status-badge-default';
  }

  typeHasSubAssets(type: string): boolean {
    if (!type) return false;
    const lower = type.toLowerCase();
    return lower === 'dpc' || lower === 'ltp';
  }

  getAvailableSubAssetTypes(): any[] {
    const defaultCodes = ['MNT', 'UPS', 'MOU', 'KBO'];
    const activeExtraCodes = this.extraSubAssetTypes;
    const activeCodes = [...defaultCodes, ...activeExtraCodes];
    return this.typeList.filter(t => activeCodes.includes(t.type.toUpperCase()));
  }

  getSelectableExtraSubAssetTypes(): any[] {
    const defaultCodes = ['MNT', 'UPS', 'MOU', 'KBO'];
    const activeExtraCodes = this.extraSubAssetTypes;
    const activeCodes = [...defaultCodes, ...activeExtraCodes];
    return this.typeList.filter(t => !activeCodes.includes(t.type.toUpperCase()));
  }

  addExtraSubAssetType(typeCode: string) {
    if (!typeCode) return;
    const codeUpper = typeCode.toUpperCase();
    if (!this.extraSubAssetTypes.includes(codeUpper)) {
      this.extraSubAssetTypes.push(codeUpper);
      const prefix = this.getSubAssetControlPrefix(codeUpper);
      
      // Register controls in addForm
      this.addForm.addControl(`has${prefix.charAt(0).toUpperCase() + prefix.slice(1)}`, this.fb.control(true));
      this.addForm.addControl(`${prefix}Brand`, this.fb.control(''));
      this.addForm.addControl(`${prefix}Model`, this.fb.control(''));
      this.addForm.addControl(`${prefix}Serial`, this.fb.control(''));
      this.addForm.addControl(`${prefix}Status`, this.fb.control('GOOD'));
      this.addForm.addControl(`${prefix}State`, this.fb.control('001'));
      this.addForm.addControl(`${prefix}Vendor`, this.fb.control(null));
      this.addForm.addControl(`${prefix}WarrantyExpiration`, this.fb.control(1));
      this.addForm.addControl(`${prefix}PurchasedValue`, this.fb.control(0));
      this.addForm.addControl(`${prefix}Watt`, this.fb.control(''));
      for (let i = 1; i <= 9; i++) {
        this.addForm.addControl(`${prefix}Detail${i}`, this.fb.control(null));
      }

      this.loadSingleSubAssetDeviceConfig(codeUpper);
      this.selectedExtraSubAssetType = null;
      this.updateSubAssetLists();
    }
  }

  removeExtraSubAssetType(typeCode: string) {
    const codeUpper = typeCode.toUpperCase();
    this.extraSubAssetTypes = this.extraSubAssetTypes.filter(c => c !== codeUpper);
    const prefix = this.getSubAssetControlPrefix(codeUpper);
    if (this.activeAddSpecTarget === prefix) {
      this.activeAddSpecTarget = 'main';
    }
    this.updateSubAssetLists();
  }

  private readonly SUB_ASSET_PREFIX_MAP: Record<string, string> = {
    'MON': 'monitor',
    'MNT': 'monitor',
    'UPS': 'ups',
    'MOU': 'mouse',
    'KBO': 'keyboard'
  };

  private readonly SUB_ASSET_CODE_MAP: Record<string, string> = {
    'monitor': 'MON',
    'ups': 'UPS',
    'mouse': 'MOU',
    'keyboard': 'KBO'
  };

  getSubAssetControlPrefix(typeCode: string): string {
    if (!typeCode) return '';
    const code = typeCode.toUpperCase();
    return this.SUB_ASSET_PREFIX_MAP[code] || code.toLowerCase();
  }

  getSubAssetCodeFromTarget(target: string): string {
    if (!target) return '';
    return this.SUB_ASSET_CODE_MAP[target] || target.toUpperCase();
  }

  loadSubAssetDeviceConfigs() {
    const defaultCodes = ['MON', 'MNT', 'UPS', 'MOU', 'KBO'];
    const activeCodes = Array.from(new Set([...defaultCodes, ...this.extraSubAssetTypes]));
    activeCodes.forEach(code => {
      this.loadSingleSubAssetDeviceConfig(code.toUpperCase());
    });
  }

  loadSingleSubAssetDeviceConfig(code: string) {
    if (!this.subAssetConfigsData[code]) {
      this.subAssetConfigsData[code] = { config: null, detailOptions: {} };
    }
    this.typeService.getDeviceConfig(code).subscribe((res: any) => {
      if (res.status && res.data) {
        this.subAssetConfigsData[code].config = res.data;
        for (let i = 1; i <= 9; i++) {
          const groupName = res.data[`detail${i}`];
          if (groupName && groupName !== 'NONE') {
            this.techSpecService.getItemsByGroup(groupName).subscribe((itemRes: any) => {
              this.subAssetConfigsData[code].detailOptions[`detail${i}`] = itemRes.status && itemRes.data ? itemRes.data : [];
            });
          } else {
            this.subAssetConfigsData[code].detailOptions[`detail${i}`] = [];
          }
        }
      }
    });
  }

  selectAddSpecTarget(target: any) {
    if (this.activeAddSpecTarget === target) {
      this.activeAddSpecTarget = 'main';
    } else {
      this.activeAddSpecTarget = target;
    }
  }

  getActiveConfig(): any {
    if (this.activeAddSpecTarget === 'main') {
      return this.addActiveDeviceConfig;
    }
    const code = this.getSubAssetCodeFromTarget(this.activeAddSpecTarget);
    return this.subAssetConfigsData[code]?.config;
  }

  getActiveDetailOptions(num: number): any[] {
    if (this.activeAddSpecTarget === 'main') {
      return this.addDetailOptions['detail' + num] || [];
    }
    const code = this.getSubAssetCodeFromTarget(this.activeAddSpecTarget);
    return this.subAssetConfigsData[code]?.detailOptions['detail' + num] || [];
  }
}
