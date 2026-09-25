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
import { TitleEmployee as TitleEmp } from '../../core/services/api.service';
import { TitleEmployeeService } from '../../core/services/title-employee.service';

@Component({
  selector: 'app-title-employee',
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
    NzInputModule
  ],
  templateUrl: './title-employee.html',
  styleUrl: './title-employee.less',
  providers: [NzModalService]
})
export class TitleEmployee implements OnInit {
  public allTitleEmployees: TitleEmp[] = [];
  public titleEmployees: TitleEmp[] = [];
  public loading = false;
  public searchText = '';
  currentFilter: any = { 
    pageIndex: 1, 
    pageSize: 100 
  };

  isVisible = false;
  isOkLoading = false;
  modalTitle = 'Add Title Employee';
  validateForm!: FormGroup;
  isEditMode = false;
  editingId: string | null = null;
  selectedEmp: TitleEmp | null = null;

  constructor(
    private titleEmployeeService: TitleEmployeeService, 
    private cdr: ChangeDetectorRef,
    private fb: FormBuilder,
    private message: NzMessageService,
    private modal: NzModalService
  ) {}

  ngOnInit() {
    this.validateForm = this.fb.group({
      employeeCode: [null, [Validators.required]],
      title: [null, [Validators.required]]
    });
    this.loadTitleEmployees();
  }

  loadTitleEmployees() {
    this.loading = true;
    this.titleEmployeeService.getAll().subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.allTitleEmployees = res.data;
        } else {
          this.allTitleEmployees = [];
        }
        this.applyFilter();
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err: any) => {
        console.error('Failed to load title employees', err);
        this.message.error('Failed to load title employees');
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  applyFilter() {
    if (!this.searchText) {
      this.titleEmployees = [...this.allTitleEmployees];
    } else {
      const lower = this.searchText.toLowerCase();
      this.titleEmployees = this.allTitleEmployees.filter(x => 
        (x.employeeCode && x.employeeCode.toLowerCase().includes(lower)) || 
        (x.title && x.title.toLowerCase().includes(lower))
      );
    }
  }

  onSearch(value: string) {
    this.searchText = value;
    this.applyFilter();
  }

  onAdd() {
    this.isEditMode = false;
    this.editingId = null;
    this.modalTitle = 'Add Title Employee';
    this.validateForm.get('employeeCode')?.enable();

    if (this.selectedEmp) {
      this.validateForm.reset({
        employeeCode: null,
        title: this.selectedEmp.title
      });
    } else {
      this.validateForm.reset();
    }
    
    this.isVisible = true;
  }

  onEdit() {
    if (!this.selectedEmp) {
      this.message.warning('Please select a record to edit');
      return;
    }
    this.isEditMode = true;
    this.editingId = this.selectedEmp.employeeCode;
    this.modalTitle = 'Edit Title Employee';
    
    this.validateForm.patchValue({
      employeeCode: this.selectedEmp.employeeCode,
      title: this.selectedEmp.title
    });
    this.validateForm.get('employeeCode')?.disable();
    this.isVisible = true;
  }

  onDelete() {
    if (!this.selectedEmp) {
      this.message.warning('Please select a record to delete');
      return;
    }
    
    this.modal.confirm({
      nzTitle: 'Are you sure delete this title employee?',
      nzContent: `<b style="color: red;">${this.selectedEmp.employeeCode} - ${this.selectedEmp.title}</b>`,
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => this.deleteEmp(this.selectedEmp!.employeeCode),
      nzCancelText: 'No'
    });
  }

  deleteEmp(id: string) {
    this.loading = true;
    this.titleEmployeeService.delete(id).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success('Deleted successfully');
          this.selectedEmp = null;
          this.loadTitleEmployees();
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
      
      const payload: TitleEmp = {
        employeeCode: formValue.employeeCode,
        title: formValue.title
      };

      if (this.isEditMode && this.editingId) {
        this.titleEmployeeService.update(this.editingId, payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Updated successfully');
              this.isVisible = false;
              this.loadTitleEmployees();
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
        this.titleEmployeeService.create(payload).subscribe({
          next: (res: any) => {
            this.isOkLoading = false;
            if (res.status) {
              this.message.success('Added successfully');
              this.isVisible = false;
              this.loadTitleEmployees();
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

  onRowSelect(emp: TitleEmp): void {
    if (this.selectedEmp?.employeeCode === emp.employeeCode) {
      this.selectedEmp = null;
    } else {
      this.selectedEmp = emp;
    }
  }
}
