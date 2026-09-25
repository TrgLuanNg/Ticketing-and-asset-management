import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormsModule } from '@angular/forms';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzRadioModule } from 'ng-zorro-antd/radio';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { Department as Dept } from '../../core/services/api.service';
import { DepartmentService } from '../../core/services/department.service';
import { UserService } from '../../core/services/user.service';

@Component({
  selector: 'app-department',
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
    NzRadioModule,
    NzSelectModule
  ],
  templateUrl: './department.html',
  styleUrl: './department.less',
  providers: [NzModalService]
})
export class Department implements OnInit {
  public allDepartments: Dept[] = [];
  public departments: Dept[] = [];
  public loading = false;
  public searchText = '';
  currentFilter: any = { 
    pageIndex: 1, 
    pageSize: 100 
  };

  isVisible = false;
  isOkLoading = false;
  modalTitle = 'Add Department';
  validateForm!: FormGroup;
  isEditMode = false;
  editingId: string | null = null;
  selectedDept: Dept | null = null;

  constructor(
    private departmentService: DepartmentService, 
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
      id: [null, [Validators.required]],
      name: [null, [Validators.required]],
      description: [null],
      type: ['Department', [Validators.required]],
      parentId: [null]
    });
    this.validateForm.get('type')?.valueChanges.subscribe(val => {
      this.updateParentIdState(val);
    });
    this.loadDepartments();
  }

  updateParentIdState(type: string) {
    const parentIdControl = this.validateForm.get('parentId');
    if (type === 'Department' || type === 'Division' || type === 'GroupDivision') {
      parentIdControl?.enable();
    } else {
      parentIdControl?.setValue(null);
      parentIdControl?.disable();
    }
  }

  loadDepartments() {
    this.loading = true;
    this.departmentService.getAll().subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.allDepartments = res.data;
        } else {
          this.allDepartments = [];
        }
        this.applyFilter();
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err: any) => {
        console.error('Failed to load departments', err);
        this.message.error('Failed to load departments');
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  applyFilter() {
    if (!this.searchText) {
      this.departments = [...this.allDepartments];
    } else {
      const lower = this.searchText.toLowerCase();
      this.departments = this.allDepartments.filter(x => 
        (x.id && x.id.toLowerCase().includes(lower)) || 
        (x.name && x.name.toLowerCase().includes(lower)) ||
        (x.description && x.description.toLowerCase().includes(lower))
      );
    }
  }

  onSearch(value: string) {
    this.searchText = value;
    this.applyFilter();
  }

  getNextId(): string {
    if (!this.allDepartments || this.allDepartments.length === 0) {
      return '00001';
    }
    const numericIds = this.allDepartments
      .map(d => parseInt(d.id, 10))
      .filter(id => !isNaN(id));
    if (numericIds.length === 0) {
      return '00001';
    }
    const max = Math.max(...numericIds);
    return (max + 1).toString().padStart(5, '0');
  }

  onAdd() {
    if (!this.hasPermission('DepartmentModify')) {
      this.message.error('You do not have permission to modify departments.');
      return;
    }
    this.isEditMode = false;
    this.editingId = null;
    this.modalTitle = 'Add Department';
    
    const nextId = this.getNextId();

    // If a department is selected, use it as a base template (cloning)
    if (this.selectedDept) {
      let type = 'Department';
      if (this.selectedDept.office) type = 'Office';
      else if (this.selectedDept.company) type = 'Company';
      else if (this.selectedDept.vendor) type = 'Vendor';
      else if (this.selectedDept.division) type = 'Division';
      else if (this.selectedDept.groupDivision) type = 'GroupDivision';

      this.validateForm.reset({
        id: nextId,
        name: this.selectedDept.name,
        description: this.selectedDept.description,
        type: type,
        parentId: this.selectedDept.parentId
      });
      this.updateParentIdState(type);
    } else {
      this.validateForm.reset({ 
        id: nextId,
        type: 'Department',
        parentId: null
      });
      this.updateParentIdState('Department');
    }
    this.validateForm.get('id')?.disable();
    this.isVisible = true;
  }

  onEdit() {
    if (!this.hasPermission('DepartmentModify')) {
      this.message.error('You do not have permission to modify departments.');
      return;
    }
    if (!this.selectedDept) {
      this.message.warning('Please select a department to edit');
      return;
    }
    this.isEditMode = true;
    this.editingId = this.selectedDept.id;
    this.modalTitle = 'Edit Department';
    
    let type = 'Unknown';
    if (this.selectedDept.department) type = 'Department';
    else if (this.selectedDept.office) type = 'Office';
    else if (this.selectedDept.company) type = 'Company';
    else if (this.selectedDept.vendor) type = 'Vendor';
    else if (this.selectedDept.division) type = 'Division';
    else if (this.selectedDept.groupDivision) type = 'GroupDivision';

    this.validateForm.patchValue({
      id: this.selectedDept.id,
      name: this.selectedDept.name,
      description: this.selectedDept.description,
      type: type,
      parentId: this.selectedDept.parentId
    });
    this.updateParentIdState(type);
    this.validateForm.get('id')?.disable();
    this.isVisible = true;
  }

  onDelete() {
    if (!this.selectedDept) {
      this.message.warning('Please select a department to delete');
      return;
    }
    
    this.modal.confirm({
      nzTitle: 'Are you sure delete this department?',
      nzContent: `<b style="color: red;">${this.selectedDept.name}</b>`,
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => this.deleteDepartment(this.selectedDept!.id),
      nzCancelText: 'No'
    });
  }

  deleteDepartment(id: string) {
    this.loading = true;
    this.departmentService.delete(id).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success('Deleted successfully');
          this.selectedDept = null;
          this.loadDepartments();
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
      
      const payload: Dept = {
        id: formValue.id,
        name: formValue.name,
        description: formValue.description,
        department: formValue.type === 'Department',
        office: formValue.type === 'Office',
        company: formValue.type === 'Company',
        vendor: formValue.type === 'Vendor',
        division: formValue.type === 'Division',
        groupDivision: formValue.type === 'GroupDivision',
        parentId: formValue.parentId || null
      };

      if (this.isEditMode && this.editingId) {
        this.departmentService.update(this.editingId, payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Updated successfully');
              this.isVisible = false;
              this.loadDepartments();
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
        this.departmentService.create(payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Added successfully');
              this.isVisible = false;
              this.loadDepartments();
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

  get parentOptions(): Dept[] {
    const currentId = this.validateForm?.get('id')?.value;
    const currentType = this.validateForm?.get('type')?.value;
    let list = this.allDepartments.filter(x => x.id !== currentId);
    if (currentType === 'Department') {
      list = list.filter(x => x.division === true);
    } else if (currentType === 'Division') {
      list = list.filter(x => x.groupDivision === true);
    }
    return list;
  }

  getTypeLabel(dept: Dept): string {
    if (dept.department) return 'Department';
    if (dept.office) return 'Office';
    if (dept.company) return 'Company';
    if (dept.vendor) return 'Vendor';
    if (dept.division) return 'Division';
    if (dept.groupDivision) return 'Group Division';
    return 'Unknown';
  }

  onRowSelect(dept: Dept): void {
    if (this.selectedDept?.id === dept.id) {
      this.selectedDept = null;
    } else {
      this.selectedDept = dept;
    }
  }
}
