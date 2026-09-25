import { Component, EventEmitter, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup } from '@angular/forms';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzIconModule } from 'ng-zorro-antd/icon';

@Component({
  selector: 'app-advanced-filter',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzButtonModule,
    NzGridModule,
    NzIconModule
  ],
  template: `
    <form nz-form nzLayout="vertical" [formGroup]="filterForm" (ngSubmit)="onSearch()" class="filter-wrapper">
      <div nz-row nzType="flex" nzAlign="bottom" [nzGutter]="[24, 24]">
        
        <div nz-col [nzFlex]="1">
          <nz-form-item class="mb-0">
            <nz-form-label>Từ khóa tìm kiếm</nz-form-label>
            <nz-form-control>
              <nz-input-group [nzPrefix]="prefixIconSearch">
                <input nz-input formControlName="keyword" placeholder="Nhập mã, tên tài sản..." />
              </nz-input-group>
              <ng-template #prefixIconSearch>
                <span nz-icon nzType="search"></span>
              </ng-template>
            </nz-form-control>
          </nz-form-item>
        </div>

        <div nz-col nzFlex="280px">
          <nz-form-item class="mb-0">
            <nz-form-label>Trạng thái</nz-form-label>
            <nz-form-control>
              <nz-select formControlName="status" nzAllowClear nzPlaceHolder="Tất cả trạng thái">
                <nz-option nzValue="GOOD" nzLabel="GOOD"></nz-option>
                <nz-option nzValue="NORMAL" nzLabel="NORMAL"></nz-option>
                <nz-option nzValue="BAD" nzLabel="BAD"></nz-option>
                <nz-option nzValue="BROKEN" nzLabel="BROKEN"></nz-option>
              </nz-select>
            </nz-form-control>
          </nz-form-item>
        </div>

        <div nz-col nzFlex="auto">
          <div class="action-buttons">
            <button nz-button nzType="primary" type="submit">
              <span nz-icon nzType="filter"></span> Lọc
            </button>
            <button nz-button nzType="default" type="button" (click)="onReset()">
              <span nz-icon nzType="reload"></span> Đặt lại
            </button>
          </div>
        </div>

      </div>
    </form>
  `,
  styles: [`
    .filter-wrapper {
      padding: 20px 24px;
      background-color: #ffffff; /* Nền trắng sáng sủa hơn */
      border: 1px solid #e8e8e8;
      border-radius: 8px;
      margin-bottom: 24px;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04); /* Đổ bóng nhẹ giúp form nổi bật */
    }

    .mb-0 {
      margin-bottom: 0 !important;
    }

    .action-buttons {
      display: flex;
      gap: 12px;
      height: 32px;
    }
  `]
})
export class AdvancedFilter {
  @Output() search = new EventEmitter<any>();
  filterForm: FormGroup;

  constructor(private fb: FormBuilder) {
    this.filterForm = this.fb.group({
      keyword: [''],
      status: [null]
    });
  }

  onSearch() {
    const rawValue = this.filterForm.value;
    const submitData = {
      ...rawValue,
      keyword: rawValue.keyword ? rawValue.keyword.trim() : ''
    };
    this.search.emit(submitData);
  }

  onReset() {
    this.filterForm.reset({
      keyword: '',
      status: null
    });
    this.search.emit(this.filterForm.value);
  }
}