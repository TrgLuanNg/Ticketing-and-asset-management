import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormsModule } from '@angular/forms';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { AssetType } from '../../core/services/api.service';
import { TypeService } from '../../core/services/type.service';
import { TechSpecService } from '../../core/services/tech-spec.service';
import { UserService } from '../../core/services/user.service';

@Component({
  selector: 'app-type',
  standalone: true,
  imports: [
    CommonModule, 
    ReactiveFormsModule,
    FormsModule,
    NzButtonModule, 
    NzIconModule, 
    NzTableModule,
    NzModalModule,
    NzFormModule,
    NzInputModule,
    NzCardModule,
    NzSelectModule,
    NzGridModule
  ],
  templateUrl: './type.html',
  styleUrl: './type.less',
  providers: [NzModalService]
})
export class Type implements OnInit {
  public allTypes: AssetType[] = [];
  public types: AssetType[] = [];
  public loading = false;
  public searchText = '';
  currentFilter: any = { 
    pageIndex: 1, 
    pageSize: 100 
  };

  isVisible = false;
  isOkLoading = false;
  modalTitle = 'Add Asset Type';
  validateForm!: FormGroup;
  isEditMode = false;
  editingId: string | null = null;
  selectedType: AssetType | null = null;

  // New properties for Device Config (Tech Specs Slots)
  public configRows: any[] = [];
  public displayConfigRows: any[] = [];
  public loadingConfig = false;
  public isConfigModalVisible = false;
  public isConfigOkLoading = false;
  public configForm!: FormGroup;
  public techSpecGroups: any[] = [];
  public deviceConfigExists = false;

  public slots = [
    { label: 'Detail 1', controlName: 'detail1' },
    { label: 'Detail 2', controlName: 'detail2' },
    { label: 'Detail 3', controlName: 'detail3' },
    { label: 'Detail 4', controlName: 'detail4' },
    { label: 'Detail 5', controlName: 'detail5' },
    { label: 'Detail 6', controlName: 'detail6' },
    { label: 'Detail 7', controlName: 'detail7' },
    { label: 'Detail 8', controlName: 'detail8' },
    { label: 'Detail 9', controlName: 'detail9' }
  ];

  constructor(
    private typeService: TypeService, 
    private techSpecService: TechSpecService,
    private cdr: ChangeDetectorRef,
    private fb: FormBuilder,
    private message: NzMessageService,
    private modal: NzModalService,
    private userService: UserService
  ) {}

  hasPermission(permission: string): boolean {
    return this.userService.hasPermission(permission);
  }

  ngOnInit() {
    this.validateForm = this.fb.group({
      type: [null, [Validators.required]],
      typeName: [null, [Validators.required]],
      vn: [null]
    });

    this.configForm = this.fb.group({
      detail1: ['NONE'],
      detail2: ['NONE'],
      detail3: ['NONE'],
      detail4: ['NONE'],
      detail5: ['NONE'],
      detail6: ['NONE'],
      detail7: ['NONE'],
      detail8: ['NONE'],
      detail9: ['NONE']
    });

    this.loadTypes();
    this.loadTechSpecGroups();
  }

  loadTypes() {
    this.loading = true;
    this.typeService.getAll().subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.allTypes = res.data;
        } else {
          this.allTypes = [];
        }
        this.applyFilter();
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err: any) => {
        console.error('Failed to load types', err);
        this.message.error('Failed to load types');
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  loadTechSpecGroups() {
    this.techSpecService.getAll().subscribe({
      next: (res: any) => {
        this.techSpecGroups = res.status && res.data ? res.data : [];
      },
      error: (err) => {
        console.error('Failed to load spec groups', err);
      }
    });
  }

  applyFilter() {
    if (!this.searchText) {
      this.types = [...this.allTypes];
    } else {
      const lower = this.searchText.toLowerCase();
      this.types = this.allTypes.filter(x => 
        (x.type && x.type.toLowerCase().includes(lower)) || 
        (x.typeName && x.typeName.toLowerCase().includes(lower)) ||
        (x.vn && x.vn.toLowerCase().includes(lower))
      );
    }
  }

  onSearch(value: string) {
    this.searchText = value;
    this.applyFilter();
  }

  onAdd() {
    if (!this.hasPermission('TypeModify')) {
      this.message.error('You do not have permission to modify asset types.');
      return;
    }
    this.isEditMode = false;
    this.editingId = null;
    this.modalTitle = 'Add Asset Type';
    this.validateForm.get('type')?.enable();

    if (this.selectedType) {
      this.validateForm.reset({
        type: null,
        typeName: this.selectedType.typeName,
        vn: this.selectedType.vn
      });
    } else {
      this.validateForm.reset();
    }
    
    this.isVisible = true;
  }

  onEdit() {
    if (!this.hasPermission('TypeModify')) {
      this.message.error('You do not have permission to modify asset types.');
      return;
    }
    if (!this.selectedType) {
      this.message.warning('Please select an asset type to edit');
      return;
    }
    this.isEditMode = true;
    this.editingId = this.selectedType.type;
    this.modalTitle = 'Edit Asset Type';
    
    this.validateForm.patchValue({
      type: this.selectedType.type,
      typeName: this.selectedType.typeName,
      vn: this.selectedType.vn
    });
    this.validateForm.get('type')?.disable();
    this.isVisible = true;
  }

  onDelete() {
    if (!this.selectedType) {
      this.message.warning('Please select an asset type to delete');
      return;
    }
    
    this.modal.confirm({
      nzTitle: 'Are you sure delete this asset type?',
      nzContent: `<b style="color: red;">${this.selectedType.typeName || this.selectedType.type}</b>`,
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => this.deleteType(this.selectedType!.type),
      nzCancelText: 'No'
    });
  }

  deleteType(id: string) {
    this.loading = true;
    this.typeService.delete(id).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success('Deleted successfully');
          this.selectedType = null;
          this.configRows = [];
          this.loadTypes();
        } else {
          this.message.error(res.message || 'Delete failed');
          this.loading = false;
        }
      },
      error: (err) => {
        console.error(err);
        this.message.error('Delete failed');
        this.loading = false;
      }
    });
  }

  handleCancel(): void {
    this.isVisible = false;
  }

  handleOk(): void {
    if (this.validateForm.valid) {
      this.isOkLoading = true;
      const formValue = this.validateForm.getRawValue();
      
      const payload: AssetType = {
        type: formValue.type,
        typeName: formValue.typeName,
        vn: formValue.vn
      };

      if (this.isEditMode && this.editingId) {
        this.typeService.update(this.editingId, payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Updated successfully');
              this.isVisible = false;
              this.loadTypes();
            } else {
              this.message.error(res.message || 'Update failed');
            }
          },
          error: (err) => {
            this.isOkLoading = false;
            console.error(err);
            this.message.error('Update failed');
          }
        });
      } else {
        this.typeService.create(payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Added successfully');
              this.isVisible = false;
              this.loadTypes();
            } else {
              this.message.error(res.message || 'Add failed');
            }
          },
          error: (err) => {
            this.isOkLoading = false;
            console.error(err);
            this.message.error('Add failed');
          }
        });
      }
    } else {
      Object.values(this.validateForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  onRowSelect(type: AssetType): void {
    if (this.selectedType?.type === type.type) {
      this.selectedType = null;
      this.configRows = [];
    } else {
      this.selectedType = type;
      this.loadDeviceConfig(type.type);
    }
  }

  loadDeviceConfig(typeId: string) {
    this.loadingConfig = true;
    this.typeService.getDeviceConfig(typeId).subscribe({
      next: (res: any) => {
        this.loadingConfig = false;
        if (res.status && res.data) {
          this.deviceConfigExists = true;
          this.buildConfigRows(res.data);
        } else {
          this.deviceConfigExists = false;
          this.buildConfigRows(null);
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        console.error('Failed to load device config', err);
        this.loadingConfig = false;
        this.deviceConfigExists = false;
        this.buildConfigRows(null);
        this.cdr.detectChanges();
      }
    });
  }

  buildConfigRows(data: any) {
    this.configRows = this.slots.map(s => {
      return {
        name: s.label,
        controlName: s.controlName,
        value: data ? data[s.controlName] : 'NONE'
      };
    });
    this.displayConfigRows = this.configRows.filter(r => r.value && r.value !== 'NONE');
  }

  onEditConfig() {
    if (!this.hasPermission('TypeModify')) {
      this.message.error('You do not have permission to modify asset types.');
      return;
    }
    if (!this.selectedType) return;
    
    const patchObj: any = {};
    this.configRows.forEach(row => {
      patchObj[row.controlName] = row.value || 'NONE';
    });
    this.configForm.patchValue(patchObj);
    this.isConfigModalVisible = true;
  }

  handleConfigOk() {
    if (!this.selectedType) return;
    
    this.isConfigOkLoading = true;
    const formVals = this.configForm.value;
    
    const payload = {
      device: this.selectedType.type,
      ...formVals
    };

    if (this.deviceConfigExists) {
      this.typeService.saveDeviceConfig(this.selectedType.type, payload).subscribe({
        next: (res: any) => {
          this.isConfigOkLoading = false;
          if (res.status) {
            this.message.success('Configured successfully');
            this.isConfigModalVisible = false;
            this.loadDeviceConfig(this.selectedType!.type);
          } else {
            this.message.error(res.message || 'Configuration failed');
          }
        },
        error: (err) => {
          this.isConfigOkLoading = false;
          console.error(err);
          this.message.error('Configuration failed');
        }
      });
    } else {
      this.typeService.createDeviceConfig(payload).subscribe({
        next: (res: any) => {
          this.isConfigOkLoading = false;
          if (res.status) {
            this.message.success('Configured successfully');
            this.isConfigModalVisible = false;
            this.loadDeviceConfig(this.selectedType!.type);
          } else {
            this.message.error(res.message || 'Configuration failed');
          }
        },
        error: (err) => {
          this.isConfigOkLoading = false;
          console.error(err);
          this.message.error('Configuration failed');
        }
      });
    }
  }
}
