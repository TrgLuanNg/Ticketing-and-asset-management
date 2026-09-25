import { Component, OnInit, ChangeDetectorRef, ViewChild, ElementRef, HostListener, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { DomSanitizer, SafeResourceUrl } from '@angular/platform-browser';
import { HttpClient } from '@angular/common/http';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzDatePickerModule } from 'ng-zorro-antd/date-picker';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzDrawerModule } from 'ng-zorro-antd/drawer';
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
import { SoftwareService, Software } from '../../core/services/software.service';
import { LookupCacheService } from '../../core/services/lookup-cache.service';
import saveAs from 'file-saver';
import { ActivatedRoute, Router } from '@angular/router';
import { Subject, EMPTY } from 'rxjs';
import { debounceTime, distinctUntilChanged, switchMap, tap } from 'rxjs/operators';

@Component({
  selector: 'app-asset-detail',
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
    NzDrawerModule,
    NzModalModule,
    NzTableModule,
    NzDropDownModule,
    NzCheckboxModule,
    NzCollapseModule,
    NzStepsModule,
    NzAutocompleteModule
  ],
  templateUrl: './asset-detail.html',
  styleUrl: './asset-detail.less',
  providers: [NzModalService],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class AssetDetail implements OnInit {
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
  public officeList: any[] = [];
  public companyList: any[] = [];
  public vendorList: any[] = [];
  public stateList: any[] = [];
  public divisionList: any[] = [];
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

  isPreviewVisible = false;
  previewTitle = '';
  previewUrl: SafeResourceUrl | null = null;
  rawPreviewUrl = '';

  public softwareRowData: Software[] = [];
  public isSoftwareLoaded = false;
  public isSoftwareLoading = false;
  public selectedSoftware: Software | null = null;
  public isSoftwareModalVisible = false;
  public softwareModalTitle = 'Thêm phần mềm';
  public isSoftwareSaving = false;
  public softwareForm!: FormGroup;
  public softwareEditMode = false;

  

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

  addDetailOptions: { [key: string]: any[] } = {};
  orgRawList: any[] = [];

  typeList: any[] = [];
  extraSubAssetTypes: string[] = [];
  availableSubAssetTypes: any[] = [];
  loadedSubAssets: any[] = [];
  isAddSubAssetModalVisible = false;
  unassignedSubAssetsList: any[] = [];
  subAssetSearchText: string = '';
  unassignedPageIndex: number = 1;
  unassignedPageSize: number = 5;
  unassignedTotalRecords: number = 0;
  isUnassignedLoading = false;
  selectedUnassignedSubAssetCode: string | null = null;
  isSubAssetLinkingLoading = false;


  activeSpecTarget: string = 'main';
  activeDeviceConfig: any = null;

  detailOptions: { [key: string]: any[] } = {};

  updateSubAssetLists() {
    const defaultCodes = ['MNT', 'UPS', 'MOU', 'KBO'];
    const activeCodes = [...defaultCodes, ...this.extraSubAssetTypes];
    
    this.availableSubAssetTypes = this.typeList.filter(t => 
      activeCodes.includes(t.type?.toUpperCase())
    );
  }

  trackByType(index: number, item: any): string {
    return item.type;
  }

  isDetailDrawerVisible = true;
  isPrintVisible = false;
  isHandoverVisible = false;
  isProblemVisible = false;
  isEditingDetails = false;
  activeAddSpecTarget: 'main' | 'monitor' | 'ups' | 'mouse' | 'keyboard' = 'main';
  subAssetConfigsData: { [typeCode: string]: { config: any, detailOptions: any } } = {
    'MNT': { config: null, detailOptions: {} },
    'UPS': { config: null, detailOptions: {} },
    'MOU': { config: null, detailOptions: {} },
    'KBO': { config: null, detailOptions: {} }
  };
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
    private userService: UserService,
    private softwareService: SoftwareService,
    private sanitizer: DomSanitizer,
    private http: HttpClient,
    private cdr: ChangeDetectorRef,
    private message: NzMessageService,
    private modalService: NzModalService,
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private lookupCacheService: LookupCacheService
  ) {
    this.createForms();
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

  ngOnInit() {
    this.loadMetadata();

    const user = this.getCurrentUser();
    const areaStr = user ? (user.area || user.Area) : '';
    if (user && areaStr) {
      this.allowedAreas = areaStr.split(',')
        .map((a: string) => a.trim())
        .map((a: string) => a.charAt(0).toUpperCase() + a.slice(1).toLowerCase())
        .filter((a: string) => a.length > 0 && a.toUpperCase() !== 'VN');
    } else {
      this.allowedAreas = [];
    }

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

    this.route.paramMap.subscribe(params => {
      const code = params.get('unitCode');
      if (code) {
        this.loadAsset(code);
      }
    });
  }

  loadMetadata() {
    this.lookupCacheService.ensureLoaded().subscribe(() => {
      this.orgRawList = this.lookupCacheService.orgRawList();
      this.departmentList = this.lookupCacheService.departmentList();
      this.officeList = this.lookupCacheService.officeList();
      this.companyList = this.lookupCacheService.companyList();
      this.vendorList = this.lookupCacheService.vendorList();
      this.divisionList = this.lookupCacheService.divisionList();
      this.groupDepList = this.lookupCacheService.groupDepList();
      this.stateList = this.lookupCacheService.stateList();
      this.typeList = this.lookupCacheService.typeList();
      this.initializeFilterFields();
      this.updateSubAssetLists();
      this.cdr.markForCheck();
    });
  }

  loadAsset(code: string): void {
    this.loading = true;
    this.cdr.markForCheck();

    this.assetService.searchGrid({
      pageIndex: 1,
      pageSize: 1,
      orderBy: 'PurchasedDate desc',
      advancedFilters: [
        { field: 'UnitCode', operator: '==', value: code }
      ]
    }).subscribe({
      next: (res: any) => {
        const data = res.status && res.data ? res.data : (Array.isArray(res) ? res : []);
        const asset = data.length > 0 ? data[0] : null;
        if (asset) {
          this.showAssetDetails(asset);
        } else {
          this.selectedAsset = null;
        }
        this.loading = false;
        this.cdr.markForCheck();
      },
      error: () => {
        this.loading = false;
        this.cdr.markForCheck();
      }
    });
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
  }

  onSearch(value: string) {
    this.onSearchChange(value);
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

  get drawerHeaderTitle(): string {
    return this.selectedAsset ? `Asset Details: ${this.selectedAsset.unitCode}` : 'Asset Details';
  }

  closeDetailDrawer() {
    this.router.navigate(['/assets']);
  }

  @HostListener('window:keydown', ['$event'])
  handleKeyDown(event: KeyboardEvent) {
    if (event.altKey && event.key === 'ArrowLeft') {
      if (this.selectedAsset) {
        event.preventDefault();
        this.closeDetailDrawer();
      }
    }
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
      emailTransfer: [''],
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

    this.softwareForm = this.fb.group({
      software: ['', [Validators.required]],
      version: ['']
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

    this.assetForm.get('type')?.valueChanges.subscribe(val => {
      if (val && !this.selectedAsset) {
        const typeUpper = val.toUpperCase();
        if (typeUpper === 'DPC') {
          this.subAssetForm.patchValue({
            hasMonitor: true,
            hasUps: true,
            hasMouse: true,
            hasKeyboard: true
          });
        } else if (typeUpper === 'LTP') {
          this.subAssetForm.patchValue({
            hasMonitor: true,
            hasUps: false,
            hasMouse: true,
            hasKeyboard: true
          });
        } else {
          this.subAssetForm.patchValue({
            hasMonitor: false,
            hasUps: false,
            hasMouse: false,
            hasKeyboard: false
          });
        }
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
    if (formName === 'asset') form = this.assetForm;
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
    if (formName === 'asset') form = this.assetForm;
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

    this.assetService.searchGrid(filterPayload).subscribe({
      next: (response: any) => {
        if (response.status && response.data) {
           this.rowData = [...response.data]; 
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

  onSelectAsset(data: Asset) {
    this.router.navigate(['/assets', data.unitCode]);
  }



  showAssetDetails(data: Asset) {
    this.selectedAsset = data;
    this.selectedHandover = null;
    this.selectedProblem = null;
    this.softwareRowData = [];
    this.isSoftwareLoaded = false;
    this.selectedSoftware = null;
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
        const subAssets: any[] = res.data;
        this.loadedSubAssets = subAssets;
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
        this.cdr.markForCheck();
      }
      
      this.onSpecTargetChange('main');
    });

    this.isDetailDrawerVisible = true;
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
      this.cdr.markForCheck();
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
              this.cdr.markForCheck();
            });
          } else {
            this.detailOptions[`detail${i}`] = [];
          }
        }
        this.cdr.markForCheck();
      } else {
        this.activeDeviceConfig = null;
        this.cdr.markForCheck();
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

  onSelectHandover(data: any) {
    if (this.selectedHandover?.id === data.id) {
      this.selectedHandover = null;
    } else {
      this.selectedHandover = data;
    }
  }

  onSelectProblem(data: any) {
    if (this.selectedProblem?.id === data.id) {
      this.selectedProblem = null;
    } else {
      this.selectedProblem = data;
    }
  }

  onViewProblem() {
    if (!this.selectedProblem) {
      this.message.warning('Please select a problem report record from the list.');
      return;
    }
    const url = `${this.problemReportService.url}/export/${this.selectedProblem.id}`;
    this.http.get(url, { responseType: 'blob' }).subscribe({
      next: (blob: Blob) => {
        if (this.rawPreviewUrl && this.rawPreviewUrl.startsWith('blob:')) {
          URL.revokeObjectURL(this.rawPreviewUrl);
        }
        const blobUrl = URL.createObjectURL(blob);
        this.rawPreviewUrl = blobUrl;
        this.previewUrl = this.sanitizer.bypassSecurityTrustResourceUrl(blobUrl);
        this.previewTitle = `Xem trước phiếu sự cố - ID: ${this.selectedProblem.id}`;
        this.isPreviewVisible = true;
        this.cdr.markForCheck();
      },
      error: (err) => {
        console.error('Error fetching problem PDF:', err);
        this.message.error('Failed to load PDF preview.');
      }
    });
  }

  loadRelatedData(unitCode: string, fromDate?: Date | null, toDate?: Date | null) {
    const filter: any = {
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'unitCode', operator: '==', value: unitCode }]
    };

    const viewer = this.isViewer() ? this.getCurrentUser() : null;
    const viewerId = viewer ? (viewer.userId || viewer.userID || '') : '';

    this.handoverService.search(filter).subscribe((res: any) => {
      let data = res.status && res.data ? res.data : [];
      let viewerCutoff: Date | null = null;

      if (viewerId && data.length > 0) {
        const viewerHandovers = data.filter((h: any) => {
          const userB = h.userTransfer || h.UserTransfer || '';
          return userB === viewerId;
        });
        if (viewerHandovers.length > 0) {
          viewerCutoff = viewerHandovers.reduce((min: Date | null, h: any) => {
            const d = h.handoverDate ? new Date(h.handoverDate) : null;
            return d && (!min || d < min) ? d : min;
          }, null as Date | null);
          data = data.filter((h: any) => h.handoverDate && new Date(h.handoverDate) >= viewerCutoff!);
        } else {
          data = [];
        }
      }

      data.sort((a: any, b: any) => {
        const timeA = a.handoverDate ? new Date(a.handoverDate).getTime() : 0;
        const timeB = b.handoverDate ? new Date(b.handoverDate).getTime() : 0;
        if (timeA !== timeB) return timeB - timeA;
        return (b.id || '').localeCompare(a.id || '');
      });
      this.handoverRowData = data;



      const cutoff = viewerCutoff;

      this.problemReportService.search(filter).subscribe((res2: any) => {
        let pData = res2.status && res2.data ? res2.data : [];
        if (cutoff) {
          pData = pData.filter((p: any) => p.reportDate && new Date(p.reportDate) >= cutoff);
        }
        pData.sort((a: any, b: any) => {
          const timeA = a.reportDate ? new Date(a.reportDate).getTime() : 0;
          const timeB = b.reportDate ? new Date(b.reportDate).getTime() : 0;
          if (timeA !== timeB) return timeB - timeA;
          return (b.id || '').localeCompare(a.id || '');
        });
        this.problemRowData = pData;
      });
    });

    if (!viewerId) {
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
    }

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
      filter.advancedFilters.push({ field: 'dateModified', operator: '<=', value: toDate.toISOString() });
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
    this.initializeFilterFields();
  }

  initializeFilterFields() {
    this.filterFields = [
      { field: 'Type', fieldName: 'Type', dataType: 'multiselect', selectDatas: this.typeList.map(t => ({ id: t.type, text: `${t.typeName} (${t.type})` })) },
      { field: 'UnitCode', fieldName: 'Unit Code', dataType: 'string' },
      { field: 'UserName', fieldName: 'User', dataType: 'string' },
      { field: 'Department', fieldName: 'Department', dataType: 'dropdown', selectDatas: this.departmentList.map(d => ({ id: d.id, text: d.name })) },
      { field: 'Office', fieldName: 'Office', dataType: 'dropdown', selectDatas: this.officeList.map(o => ({ id: o.id, text: o.name })) },
      { field: 'Company', fieldName: 'Company', dataType: 'dropdown', selectDatas: this.companyList.map(c => ({ id: c.id, text: c.name })) },
      { field: 'Area', fieldName: 'Area', dataType: 'dropdown', selectDatas: this.allowedAreas.map(a => ({ id: a, text: a })) },
      { field: 'HRCode', fieldName: 'HR Code', dataType: 'string' },
      { field: 'Email', fieldName: 'Email', dataType: 'string' },
      { field: 'Status', fieldName: 'Status', dataType: 'dropdown', selectDatas: [
        { id: 'GOOD', text: 'GOOD' },
        { id: 'NORMAL', text: 'NORMAL' },
        { id: 'BAD', text: 'BAD' },
        { id: 'BROKEN', text: 'BROKEN' }
      ]},
      { field: 'State', fieldName: 'State', dataType: 'multiselect', selectDatas: this.stateList.map(s => ({ id: s.id, text: s.description || s.state })) },
      { field: 'Vendor', fieldName: 'Vendor', dataType: 'string' },
      { field: 'Brand', fieldName: 'Brand', dataType: 'string' },
      { field: 'Model', fieldName: 'Model', dataType: 'string' },
      { field: 'Serial', fieldName: 'Serial', dataType: 'string' },
      { field: 'Division', fieldName: 'Division', dataType: 'dropdown', selectDatas: this.divisionList.map(d => ({ id: d.id, text: d.name })) },
      { field: 'GroupDep', fieldName: 'Group Dep', dataType: 'dropdown', selectDatas: this.groupDepList.map(g => ({ id: g.id, text: g.name })) },
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
          .print-qr { width:1.2cm; height:1.2cm; position:absolute; top:0.4cm; left:0.4cm; } 
          .print-info { position:absolute; top:0.4cm; left:1.8cm; right:0.4cm; font-size:8px; line-height:1.2; word-break:break-all; color:#000000; text-align:left; margin:0; padding:0; } 
          .print-unit-code { font-size:9.5px; font-weight:bold; margin:0 0 2px 0; padding:0; line-height:1; color:#000000; } 
          .print-info-item { margin-bottom:1px; }
        </style>
      `;
      printWindow.document.write(styles);
      printWindow.document.write('</head><body>');

      selectedAssets.forEach(asset => {
        const qrDataString = `Unit ${asset.unitCode || ''}| Vendor ${asset.vendor || ''}| Serial ${asset.serial || ''}| PurchasedDate ${asset.purchasedDate ? new Date(asset.purchasedDate).toLocaleDateString('en-GB') : ''}| Status ${asset.status || ''}`;
        const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(qrDataString)}`;
        const purDateStr = asset.purchasedDate ? `Pur. Date: ${new Date(asset.purchasedDate).toLocaleDateString('en-GB')}` : 'Pur. Date: N/A';
        const vendorStr = `Vendor: ${asset.vendor || 'N/A'}`;
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

  onEdit() {
    if (!this.hasPermission('AssetsModify')) {
      this.message.error('You do not have permission to modify assets.');
      return;
    }
    this.startEditDetails();
  }

  startEditDetails() {
    if (!this.hasPermission('AssetsModify')) {
      this.message.error('You do not have permission to modify assets.');
      return;
    }
    this.isEditingDetails = true;
    this.assetForm.enable();
    this.specForm.enable();
    this.assetForm.get('unitCode')?.disable();
    this.assetForm.get('ad')?.disable();
    this.assetForm.get('me')?.disable();
    this.assetForm.get('tm')?.disable();
  }

  cancelEditDetails() {
    this.isEditingDetails = false;
    if (this.selectedAsset) {
      this.showAssetDetails(this.selectedAsset);
    }
  }

  onTabIndexChange(index: number) {
    this.activeDetailTabIndex = index;
    if (index !== 0 && this.isEditingDetails) {
      this.isEditingDetails = false;
      if (this.selectedAsset) {
        this.assetForm.patchValue({
          ...this.selectedAsset,
          purchasedDate: this.selectedAsset.purchasedDate ? new Date(this.selectedAsset.purchasedDate) : null,
          usageDate: this.selectedAsset.usageDate ? new Date(this.selectedAsset.usageDate) : null,
          recordDate: this.selectedAsset.recordDate ? new Date(this.selectedAsset.recordDate) : null
        });
        this.assetForm.disable();
        this.specForm.disable();
      }
    }
    this.cdr.detectChanges();
  }

  saveEditDetails() {
    this.saveActiveSpecValues();
    if (this.selectedAsset && this.assetForm.valid) {
      const rawFormValues = this.assetForm.getRawValue();
      const updatedAsset = {
        ...this.selectedAsset,
        ...rawFormValues
      };
      this.assetService.update(this.selectedAsset.unitCode, updatedAsset).subscribe({
        next: (res: any) => {
          if (res.status) {
             this.message.success(res.message || 'Asset updated successfully');
             this.isEditingDetails = false;
             this.assetForm.disable();
             this.specForm.disable();
             const latestData = res.data || updatedAsset;
             this.selectedAsset = { ...latestData };
             this.showAssetDetails(this.selectedAsset);
             this.assetService.assetUpdated$.next(this.selectedAsset);
          } else {
             this.message.error(res.message || 'Failed to update asset', { nzDuration: 6000 });
          }
        },
        error: (err: any) => {
           console.error(err);
           this.message.error('An error occurred while updating the asset');
        }
      });
    }
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

  onViewHandover() {
    if (!this.selectedHandover) {
      this.message.warning('Please select a handover record from the list.');
      return;
    }
    const url = `${this.handoverService.url}/export/${this.selectedHandover.id}?unitCode=${this.selectedHandover.unitCode}`;
    this.http.get(url, { responseType: 'blob' }).subscribe({
      next: (blob: Blob) => {
        if (this.rawPreviewUrl && this.rawPreviewUrl.startsWith('blob:')) {
          URL.revokeObjectURL(this.rawPreviewUrl);
        }
        const blobUrl = URL.createObjectURL(blob);
        this.rawPreviewUrl = blobUrl;
        this.previewUrl = this.sanitizer.bypassSecurityTrustResourceUrl(blobUrl);
        this.previewTitle = `Xem trước biên bản bàn giao - ID: ${this.selectedHandover.id}`;
        this.isPreviewVisible = true;
        this.cdr.markForCheck();
      },
      error: (err) => {
        console.error('Error fetching handover PDF:', err);
        this.message.error('Failed to load PDF preview.');
      }
    });
  }

  closePreview() {
    this.isPreviewVisible = false;
    if (this.rawPreviewUrl && this.rawPreviewUrl.startsWith('blob:')) {
      URL.revokeObjectURL(this.rawPreviewUrl);
      this.rawPreviewUrl = '';
      this.previewUrl = null;
    }
  }

  printPdf() {
    const iframe = document.getElementById('pdfPreviewIframe') as HTMLIFrameElement;
    if (iframe && iframe.contentWindow) {
      iframe.contentWindow.focus();
      try {
        iframe.contentWindow.print();
      } catch (e) {
        console.error('Direct printing from iframe failed, fallback to new window:', e);
        const printWindow = window.open(this.rawPreviewUrl, '_blank');
        if (printWindow) {
          printWindow.print();
        }
      }
    } else if (this.rawPreviewUrl) {
      const printWindow = window.open(this.rawPreviewUrl, '_blank');
      if (printWindow) {
        printWindow.print();
      }
    }
  }

  onProblemReport() {
    if (!this.selectedAsset) {
      this.message.warning('Please select an asset to report a problem.');
      return;
    }
    this.problemEditMode = false;
    this.problemForm.reset();

    if (this.isViewer()) {
      this.problemForm.get('testProcedures')?.clearValidators();
      this.problemForm.get('solutions')?.clearValidators();
    } else {
      this.problemForm.get('testProcedures')?.setValidators([Validators.required]);
      this.problemForm.get('solutions')?.setValidators([Validators.required]);
    }
    this.problemForm.get('testProcedures')?.updateValueAndValidity();
    this.problemForm.get('solutions')?.updateValueAndValidity();

    this.isProblemVisible = true;
  }

  onModifyProblem() {
    if (this.isViewer()) {
      this.message.error('Viewers cannot modify problems.');
      return;
    }
    if (!this.selectedProblem) {
      this.message.warning('Please select a problem report record to modify.');
      return;
    }
    this.problemEditMode = true;
    this.problemForm.patchValue({
      problemSummary: this.selectedProblem.problemSummary,
      testProcedures: this.selectedProblem.testProcedures,
      solutions: this.selectedProblem.solutions,
      jiraCode: this.selectedProblem.jiraCode
    });
    this.isProblemVisible = true;
  }

  submitProblemReport() {
    if (this.problemForm.valid) {
      this.isProblemVisible = false;
      const user = this.getCurrentUser();
      const creatorUserId = user ? (user.userId || user.userID) : 'System';

      const payload = {
        ...this.problemForm.value,
        unitCode: this.selectedAsset.unitCode,
        userCreated: creatorUserId,
        type: this.selectedAsset?.type,
        detail1: this.selectedAsset?.detail1,
        detail2: this.selectedAsset?.detail2,
        detail3: this.selectedAsset?.detail3,
        detail4: this.selectedAsset?.detail4,
        detail5: this.selectedAsset?.detail5,
        detail6: this.selectedAsset?.detail6,
        detail7: this.selectedAsset?.detail7,
        detail8: this.selectedAsset?.detail8,
        detail9: this.selectedAsset?.detail9,
        departmentReport: this.selectedAsset?.department,
        officeReport: this.selectedAsset?.office,
        companyReport: this.selectedAsset?.company,
        areaReport: this.selectedAsset?.division,
        userReport: this.selectedAsset?.userName
      };

      if (this.problemEditMode) {
        payload.id = this.selectedProblem.id;
        payload.reportDate = this.selectedProblem.reportDate;
        payload.userCreated = this.selectedProblem.userCreated;
        this.problemReportService.update(this.selectedProblem.id, payload).subscribe({
          next: (res: any) => {
            if (res.status) {
              this.message.success('Problem report updated.');
              this.isProblemVisible = false;
              this.loadRelatedData(this.selectedAsset.unitCode);
              this.selectedProblem = null;
            } else {
              this.message.error(res.message || 'Failed to update problem report.');
            }
          },
          error: (err: any) => {
            console.error(err);
            this.message.error('An error occurred.');
          }
        });
      } else {
        this.problemReportService.create(payload).subscribe({
          next: (res: any) => {
            if (res.status) {
              this.message.success('Problem report created.');
              this.isProblemVisible = false;
              this.loadRelatedData(this.selectedAsset.unitCode);
            } else {
              this.message.error(res.message || 'Failed to create problem report.');
            }
          },
          error: (err: any) => {
            console.error(err);
            this.message.error('An error occurred.');
          }
        });
      }
    } else {
      Object.values(this.problemForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  onTransferState() {
    if (this.isViewer()) {
      this.message.error('Viewers cannot transfer assets.');
      return;
    }
    if (!this.selectedAsset) {
      this.message.warning('Please select an asset to transfer.');
      return;
    }

    let determinedForm = 'ASS';
    const state = this.selectedAsset.state;
    const currentUser = (this.selectedAsset.userName || '').toUpperCase();

    if (state === '002') {
      determinedForm = (currentUser !== 'IT' && currentUser !== '') ? 'TRF' : 'RETURN';
    } else if (state === '001') {
      determinedForm = 'ASS';
    } else {
      if (currentUser !== 'IT' && currentUser !== '') {
        determinedForm = 'RETURN';
      } else {
        determinedForm = 'ASS';
      }
    }

    this.handoverForm.patchValue({
      form: determinedForm,
      retrieveToITD: false
    }, { emitEvent: false });

    this.handleFormTypeChange(determinedForm);

    if (this.handoverFormValueChangesSub) {
      this.handoverFormValueChangesSub.unsubscribe();
    }
    this.handoverFormValueChangesSub = this.handoverForm.get('form')?.valueChanges.subscribe((val: string) => {
      this.handleFormTypeChange(val);
    });

    if (this.handoverRetrieveSub) {
      this.handoverRetrieveSub.unsubscribe();
    }
    this.handoverRetrieveSub = this.handoverForm.get('retrieveToITD')?.valueChanges.subscribe((retrieve: boolean) => {
      if (this.handoverForm.get('form')?.value === 'PRB') {
        this.handleProblemRetrieveChange(retrieve);
      }
    });

    this.isHandoverVisible = true;
  }

  private handoverFormValueChangesSub: any = null;
  private handoverRetrieveSub: any = null;

  handleFormTypeChange(val: string) {
    const asset = this.selectedAsset;
    if (!asset) return;

    const allControls = ['userTransfer', 'departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'detail1', 'detail2', 'divisionTransfer', 'groupDepTransfer', 'emailTransfer'];
    allControls.forEach(ctrl => this.handoverForm.get(ctrl)?.enable());

    if (val === 'ASS') {
      this.isTargetFieldsRequired = true;
      this.updateHandoverValidators(true);
      this.handoverForm.patchValue({
        detail1: '',
        userTransfer: '',
        departmentTransfer: null,
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: null,
        groupDepTransfer: null,
        emailTransfer: asset.email || '',
        detail2: 'GOOD'
      }, { emitEvent: false });

    } else if (val === 'LOA') {
      this.isTargetFieldsRequired = true;
      this.updateHandoverValidators(true);
      this.handoverForm.patchValue({
        detail1: '',
        userTransfer: '',
        departmentTransfer: null,
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: null,
        groupDepTransfer: null,
        emailTransfer: asset.email || '',
        detail2: 'GOOD'
      }, { emitEvent: false });

    } else if (val === 'RETURN') {
      this.isTargetFieldsRequired = false;
      this.updateHandoverValidators(false);
      this.handoverForm.patchValue({
        detail1: asset.hrCode || '',
        userTransfer: asset.userName || '',
        departmentTransfer: asset.department || '',
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: asset.division || '',
        groupDepTransfer: asset.groupDep || '',
        emailTransfer: asset.email || '',
        detail2: 'GOOD'
      }, { emitEvent: false });

      ['userTransfer', 'departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'detail1', 'divisionTransfer', 'groupDepTransfer'].forEach(ctrl => {
        this.handoverForm.get(ctrl)?.disable();
      });

    } else if (val === 'TRF') {
      this.isTargetFieldsRequired = true;
      this.updateHandoverValidators(true);
      this.handoverForm.patchValue({
        detail1: '',
        userTransfer: '',
        departmentTransfer: null,
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: null,
        groupDepTransfer: null,
        emailTransfer: asset.email || '',
        detail2: asset.status || 'GOOD'
      }, { emitEvent: false });

    } else if (val === 'PRB') {
      this.handoverForm.patchValue({
        detail2: 'BROKEN',
        emailTransfer: asset.email || '',
        retrieveToITD: false
      }, { emitEvent: false });
      this.handleProblemRetrieveChange(false);

    } else if (val === 'LOST') {
      this.isTargetFieldsRequired = false;
      this.updateHandoverValidators(false);
      this.handoverForm.patchValue({
        detail1: asset.hrCode || '',
        userTransfer: asset.userName || '',
        departmentTransfer: asset.department || '',
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: asset.division || '',
        groupDepTransfer: asset.groupDep || '',
        emailTransfer: asset.email || '',
        detail2: asset.status || 'GOOD'
      }, { emitEvent: false });

      ['userTransfer', 'departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'detail1', 'detail2', 'divisionTransfer', 'groupDepTransfer'].forEach(ctrl => {
        this.handoverForm.get(ctrl)?.disable();
      });
    }
  }

  handleProblemRetrieveChange(retrieve: boolean) {
    const asset = this.selectedAsset;
    if (!asset) return;

    if (retrieve) {
      this.isTargetFieldsRequired = false;
      this.updateHandoverValidators(false);
      this.handoverForm.patchValue({
        detail1: asset.hrCode || '',
        userTransfer: asset.userName || '',
        departmentTransfer: asset.department || '',
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: asset.division || '',
        groupDepTransfer: asset.groupDep || '',
        emailTransfer: asset.email || '',
      }, { emitEvent: false });

      ['userTransfer', 'departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'detail1', 'divisionTransfer', 'groupDepTransfer'].forEach(ctrl => {
        this.handoverForm.get(ctrl)?.disable();
      });
    } else {
      this.isTargetFieldsRequired = true;
      this.updateHandoverValidators(true);
      this.handoverForm.patchValue({
        detail1: asset.hrCode || '',
        userTransfer: asset.userName || '',
        departmentTransfer: null,
        officeTransfer: asset.office || '',
        companyTransfer: asset.company || 'YLVN',
        areaTransfer: asset.area || 'VN',
        divisionTransfer: null,
        groupDepTransfer: null,
        emailTransfer: asset.email || '',
      }, { emitEvent: false });

      ['userTransfer', 'departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'detail1', 'divisionTransfer', 'groupDepTransfer'].forEach(ctrl => {
        this.handoverForm.get(ctrl)?.enable();
      });
    }
  }

  updateHandoverValidators(isRequired: boolean) {
    const fields = ['departmentTransfer', 'officeTransfer', 'companyTransfer', 'areaTransfer', 'divisionTransfer', 'groupDepTransfer'];
    fields.forEach(field => {
      const control = this.handoverForm.get(field);
      if (control) {
        if (isRequired) {
          control.setValidators([Validators.required]);
        } else {
          control.clearValidators();
        }
        control.updateValueAndValidity({ emitEvent: false });
      }
    });
  }

  submitHandover() {
    if (this.handoverForm.valid) {
      this.isHandoverVisible = false;
      const payload = {
        ...this.handoverForm.getRawValue(),
        unitCode: this.selectedAsset.unitCode
      };

      this.handoverService.create(payload).subscribe({
        next: (res: any) => {
          if (res.status) {
            this.message.success('Handover created and Asset updated.');
            this.isHandoverVisible = false;
            this.assetService.getById(this.selectedAsset.unitCode).subscribe((assetRes: any) => {
              if (assetRes.status && assetRes.data) {
                this.showAssetDetails(assetRes.data);
                this.assetService.assetUpdated$.next(assetRes.data);
              }
            });
          } else {
            this.message.error(res.message || 'Failed to create handover.');
          }
        },
        error: (err: any) => {
          console.error(err);
          this.message.error('An error occurred.');
        }
      });
    } else {
      Object.values(this.handoverForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  onReviewLogs() {
    if (this.selectedAsset) {
      this.loadLogs(this.selectedAsset.unitCode, this.logFromDate, this.logToDate);
    }
  }

  onExportLogs() {
    if (!this.selectedAsset) return;
    
    const fromStr = this.logFromDate ? this.logFromDate.toISOString() : '';
    const toStr = this.logToDate ? this.logToDate.toISOString() : '';
    
    const url = `${this.assetsLogService.url}/export?unitCode=${this.selectedAsset.unitCode}&fromDate=${fromStr}&toDate=${toStr}`;
    window.open(url, '_blank');
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

  loadSubAssetDeviceConfigs() {
    const subCodes = ['MNT', 'UPS', 'MOU', 'KBO'];
    subCodes.forEach(code => {
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
    });
  }
  private readonly SUB_ASSET_PREFIX_MAP: Record<string, string> = {
    'MNT': 'monitor',
    'UPS': 'ups',
    'MOU': 'mouse',
    'KBO': 'keyboard'
  };

  private readonly SUB_ASSET_CODE_MAP: Record<string, string> = {
    'monitor': 'MNT',
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
    const subAsset = this.loadedSubAssets?.find(s => s.unitCode === target);
    if (subAsset) return subAsset.type;
    return this.SUB_ASSET_CODE_MAP[target] || target.toUpperCase();
  }

  getSubAssetTabLabel(sub: any): string {
    if (!sub) return '';
    return sub.unitCode || '';
  }


  trackByUnitCode(index: number, item: any): string {
    return item?.unitCode || index.toString();
  }

  openAddSubAssetModal() {
    if (!this.selectedAsset) return;
    this.isAddSubAssetModalVisible = true;
    this.selectedUnassignedSubAssetCode = null;
    this.subAssetSearchText = '';
    this.unassignedPageIndex = 1;
    this.loadUnassignedSubAssets();
  }

  loadUnassignedSubAssets() {
    this.isUnassignedLoading = true;
    this.assetService.getUnassignedSubAssets({
      pageIndex: this.unassignedPageIndex,
      pageSize: this.unassignedPageSize,
      searchText: this.subAssetSearchText
    }).subscribe({
      next: (res: any) => {
        this.isUnassignedLoading = false;
        if (res.status && res.data) {
          this.unassignedSubAssetsList = res.data;
          this.unassignedTotalRecords = res.totalRecords || 0;
        } else {
          this.unassignedSubAssetsList = [];
          this.unassignedTotalRecords = 0;
        }
        this.cdr.markForCheck();
      },
      error: () => {
        this.isUnassignedLoading = false;
        this.message.error('Failed to fetch unassigned sub-assets');
        this.cdr.markForCheck();
      }
    });
  }

  onUnassignedPageIndexChange(pageIndex: number) {
    this.unassignedPageIndex = pageIndex;
    this.loadUnassignedSubAssets();
  }

  onUnassignedSearchTextChange(text: string) {
    this.subAssetSearchText = text;
    this.unassignedPageIndex = 1;
    this.loadUnassignedSubAssets();
  }

  closeAddSubAssetModal() {
    this.isAddSubAssetModalVisible = false;
    this.selectedUnassignedSubAssetCode = null;
    this.subAssetSearchText = '';
  }



  confirmAddSubAsset() {
    if (!this.selectedAsset || !this.selectedUnassignedSubAssetCode) return;
    this.isSubAssetLinkingLoading = true;

    const subCode = this.selectedUnassignedSubAssetCode;
    const mainCode = this.selectedAsset.unitCode;

    this.assetService.updateMainAsset(subCode, mainCode).subscribe({
      next: (res: any) => {
        this.isSubAssetLinkingLoading = false;
        if (res.status) {
          this.message.success(`Sub-asset ${subCode} added to ${mainCode}`);
          this.closeAddSubAssetModal();
          this.reloadSubAssets(mainCode, subCode);
        } else {
          this.message.error(res.message || 'Failed to link sub-asset');
        }
        this.cdr.markForCheck();
      },
      error: () => {
        this.isSubAssetLinkingLoading = false;
        this.message.error('Failed to link sub-asset');
        this.cdr.markForCheck();
      }
    });
  }

  clearSubAsset(sub: any, event: Event) {
    event.stopPropagation();
    if (!sub || !sub.unitCode) return;

    this.modalService.confirm({
      nzTitle: 'Clear Sub-Asset Link?',
      nzContent: `Are you sure you want to remove ${sub.unitCode} from main asset ${this.selectedAsset?.unitCode}?`,
      nzOnOk: () => {
        this.assetService.updateMainAsset(sub.unitCode, null).subscribe({
          next: (res: any) => {
            if (res.status) {
              this.message.success(`Sub-asset ${sub.unitCode} cleared`);
              if (this.activeSpecTarget === sub.unitCode) {
                this.onSpecTargetChange('main');
              }
              if (this.selectedAsset) {
                this.reloadSubAssets(this.selectedAsset.unitCode);
              }
            } else {
              this.message.error(res.message || 'Failed to clear sub-asset');
            }
            this.cdr.markForCheck();
          },
          error: () => {
            this.message.error('Failed to clear sub-asset');
            this.cdr.markForCheck();
          }
        });
      }
    });
  }

  reloadSubAssets(mainUnitCode: string, selectUnitCodeTarget?: string) {
    this.assetService.search({
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'MainAsset', operator: '==', value: mainUnitCode }]
    }).subscribe((res: any) => {
      if (res.status && res.data) {
        this.loadedSubAssets = res.data;
        this.loadedSubAssets.forEach(sub => {
          const typeUpper = sub.type?.toUpperCase();
          const prefix = this.getSubAssetControlPrefix(typeUpper);
          const patchObj: any = {};
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
          if (this.subAssetForm) {
            this.subAssetForm.patchValue(patchObj);
          }
        });
      } else {
        this.loadedSubAssets = [];
      }
      if (selectUnitCodeTarget) {
        this.onSpecTargetChange(selectUnitCodeTarget);
      }
      this.cdr.markForCheck();
    });
  }


  isLtpOrDpc(): boolean {
    if (this.isViewer()) return false;
    if (!this.selectedAsset || !this.selectedAsset.type) return false;
    const t = this.selectedAsset.type.toUpperCase().trim();
    return t === 'LTP' || t === 'DPC';
  }

  onSoftwareTabSelect() {
    if (!this.isSoftwareLoaded && this.selectedAsset?.unitCode) {
      this.loadSoftwareData();
    }
  }

  loadSoftwareData() {
    if (!this.selectedAsset?.unitCode) return;
    this.isSoftwareLoading = true;
    const filter: any = {
      pageIndex: 1,
      pageSize: 100,
      advancedFilters: [{ field: 'unitCode', operator: '==', value: this.selectedAsset.unitCode }]
    };
    this.softwareService.search(filter).subscribe({
      next: (res: any) => {
        this.softwareRowData = res.status && res.data ? res.data : [];
        this.isSoftwareLoaded = true;
        this.isSoftwareLoading = false;
        this.cdr.markForCheck();
      },
      error: (err: any) => {
        console.error('Error loading software list:', err);
        this.isSoftwareLoading = false;
        this.message.error('Failed to load software list');
        this.cdr.markForCheck();
      }
    });
  }

  onSelectSoftware(data: Software) {
    if (this.selectedSoftware?.id === data.id) {
      this.selectedSoftware = null;
    } else {
      this.selectedSoftware = data;
    }
  }

  onAddSoftware() {
    this.softwareEditMode = false;
    this.softwareModalTitle = 'Thêm phần mềm (Add Software)';
    this.softwareForm.reset();
    this.isSoftwareModalVisible = true;
  }

  onEditSoftware() {
    if (!this.selectedSoftware) {
      this.message.warning('Vui lòng chọn phần mềm cần sửa!');
      return;
    }
    this.softwareEditMode = true;
    this.softwareModalTitle = 'Sửa phần mềm (Edit Software)';
    this.softwareForm.patchValue({
      software: this.selectedSoftware.software,
      version: this.selectedSoftware.version || ''
    });
    this.isSoftwareModalVisible = true;
  }

  onDeleteSoftware() {
    if (!this.selectedSoftware || !this.selectedSoftware.id) {
      this.message.warning('Vui lòng chọn phần mềm cần xóa!');
      return;
    }
    this.modalService.confirm({
      nzTitle: 'Xác nhận xóa',
      nzContent: `Bạn có chắc chắn muốn xóa phần mềm "${this.selectedSoftware.software}" không?`,
      nzOkText: 'Xóa',
      nzOkDanger: true,
      nzOnOk: () => {
        this.softwareService.delete(this.selectedSoftware!.id!).subscribe({
          next: () => {
            this.message.success('Xóa phần mềm thành công');
            this.selectedSoftware = null;
            this.loadSoftwareData();
          },
          error: (err: any) => {
            console.error('Error deleting software:', err);
            this.message.error('Xóa phần mềm thất bại');
          }
        });
      }
    });
  }

  onSaveSoftware() {
    if (this.softwareForm.invalid) {
      Object.values(this.softwareForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      return;
    }

    const val = this.softwareForm.value;
    this.isSoftwareSaving = true;

    if (this.softwareEditMode && this.selectedSoftware?.id) {
      const payload: Software = {
        id: this.selectedSoftware.id,
        unitCode: this.selectedAsset.unitCode,
        software: val.software,
        version: val.version
      };
      this.softwareService.update(this.selectedSoftware.id, payload).subscribe({
        next: () => {
          this.message.success('Cập nhật phần mềm thành công');
          this.isSoftwareSaving = false;
          this.isSoftwareModalVisible = false;
          this.selectedSoftware = null;
          this.loadSoftwareData();
        },
        error: (err: any) => {
          console.error('Error updating software:', err);
          this.isSoftwareSaving = false;
          this.message.error('Cập nhật phần mềm thất bại');
          this.cdr.markForCheck();
        }
      });
    } else {
      const payload: Software = {
        unitCode: this.selectedAsset.unitCode,
        software: val.software,
        version: val.version
      };
      this.softwareService.create(payload).subscribe({
        next: () => {
          this.message.success('Thêm phần mềm thành công');
          this.isSoftwareSaving = false;
          this.isSoftwareModalVisible = false;
          this.loadSoftwareData();
        },
        error: (err: any) => {
          console.error('Error adding software:', err);
          this.isSoftwareSaving = false;
          this.message.error('Thêm phần mềm thất bại');
          this.cdr.markForCheck();
        }
      });
    }
  }
}
