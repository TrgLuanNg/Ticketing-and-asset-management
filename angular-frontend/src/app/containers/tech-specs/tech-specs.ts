import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormsModule } from '@angular/forms';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzCardModule } from 'ng-zorro-antd/card';
import { TechSpecKey, TechSpecItem } from '../../core/services/api.service';
import { TechSpecService } from '../../core/services/tech-spec.service';

@Component({
  selector: 'app-tech-specs',
  standalone: true,
  imports: [
    CommonModule, 
    ReactiveFormsModule,
    FormsModule,
    NzButtonModule, 
    NzIconModule, 
    NzTableModule, 
    NzGridModule,
    NzModalModule,
    NzFormModule,
    NzInputModule,
    NzCardModule
  ],
  templateUrl: './tech-specs.html',
  styleUrl: './tech-specs.less',
  providers: [NzModalService]
})
export class TechSpecs implements OnInit {
  public techSpecKeys: TechSpecKey[] = [];
  public selectedKey: TechSpecKey | null = null;
  public techSpecItems: TechSpecItem[] = [];
  public selectedItem: TechSpecItem | null = null;
  public searchInput: string = '';
  public activeSearchQuery: string = '';
  
  public loadingKeys = false;
  public loadingItems = false;

  get filteredTechSpecKeys(): TechSpecKey[] {
    if (!this.activeSearchQuery || !this.activeSearchQuery.trim()) {
      return this.techSpecKeys;
    }
    const q = this.activeSearchQuery.trim().toLowerCase();
    return this.techSpecKeys.filter(k => 
      (k.name && k.name.toLowerCase().includes(q)) || 
      (k.descriptionTechSpecs && k.descriptionTechSpecs.toLowerCase().includes(q))
    );
  }

  onSearchKey() {
    this.activeSearchQuery = this.searchInput;
  }

  // Group Modal (Tech Spec Key)
  isGroupModalVisible = false;
  isGroupOkLoading = false;
  groupForm!: FormGroup;
  isGroupEditMode = false;
  groupModalTitle = 'Add Tech Spec Group';

  // Item Modal (Tech Spec Item)
  isItemModalVisible = false;
  isItemOkLoading = false;
  itemForm!: FormGroup;
  isItemEditMode = false;
  itemModalTitle = 'Add Tech Spec Value';

  constructor(
    private techSpecService: TechSpecService,
    private cdr: ChangeDetectorRef,
    private fb: FormBuilder,
    private message: NzMessageService,
    private modal: NzModalService
  ) {}

  ngOnInit() {
    this.groupForm = this.fb.group({
      name: [null, [Validators.required]],
      descriptionTechSpecs: [null]
    });

    this.itemForm = this.fb.group({
      id: [null],
      specGroup: [{ value: null, disabled: true }],
      item: [null, [Validators.required]],
      techSpecs: [null]
    });

    this.loadTechSpecs();
  }

  loadTechSpecs() {
    this.loadingKeys = true;
    this.techSpecService.getAll().subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.techSpecKeys = res.data;
        } else {
          this.techSpecKeys = [];
        }
        this.loadingKeys = false;
        if (this.techSpecKeys.length > 0 && !this.selectedKey) {
          this.onSelectKey(this.techSpecKeys[0]);
        }
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err) => {
        console.error('Failed to load tech spec keys', err);
        this.message.error('Failed to load groups');
        this.loadingKeys = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  onSelectKey(key: TechSpecKey) {
    this.selectedKey = key;
    this.selectedItem = null;
    this.loadingItems = true;
    this.techSpecService.getItemsByGroup(key.name).subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.techSpecItems = res.data;
        } else if (Array.isArray(res)) {
          this.techSpecItems = res;
        } else {
          this.techSpecItems = [];
        }
        this.loadingItems = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err) => {
        console.error('Failed to load items', err);
        this.message.error('Failed to load items');
        this.techSpecItems = [];
        this.loadingItems = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  onRowSelectKey(key: TechSpecKey): void {
    this.onSelectKey(key);
  }

  onRowSelectItem(item: TechSpecItem): void {
    if (this.selectedItem?.id === item.id) {
      this.selectedItem = null;
    } else {
      this.selectedItem = item;
    }
  }

  // --- Group actions ---
  onAddTech() {
    this.isGroupEditMode = false;
    this.groupModalTitle = 'Add Tech Spec Group';
    this.groupForm.reset();
    this.groupForm.get('name')?.enable();
    this.isGroupModalVisible = true;
  }

  onEditTech() {
    if (!this.selectedKey) {
      this.message.warning('Please select a group to edit');
      return;
    }
    this.isGroupEditMode = true;
    this.groupModalTitle = 'Edit Tech Spec Group';
    this.groupForm.patchValue({
      name: this.selectedKey.name,
      descriptionTechSpecs: this.selectedKey.descriptionTechSpecs
    });
    this.groupForm.get('name')?.disable();
    this.isGroupModalVisible = true;
  }

  onDeleteTech() {
    if (!this.selectedKey) {
      this.message.warning('Please select a group to delete');
      return;
    }
    this.modal.confirm({
      nzTitle: 'Are you sure delete this spec group?',
      nzContent: `<b style="color: red;">${this.selectedKey.name}</b>`,
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => {
        this.loadingKeys = true;
        this.techSpecService.delete(this.selectedKey!.name).subscribe({
          next: (res: any) => {
            if (res.status) {
              this.message.success('Group deleted');
              this.selectedKey = null;
              this.loadTechSpecs();
            } else {
              this.message.error(res.message || 'Delete failed');
              this.loadingKeys = false;
            }
          },
          error: (err) => {
            console.error(err);
            this.message.error('Delete failed');
            this.loadingKeys = false;
          }
        });
      }
    });
  }

  handleGroupOk() {
    if (this.groupForm.valid) {
      this.isGroupOkLoading = true;
      const formVal = this.groupForm.getRawValue();
      const payload: TechSpecKey = {
        name: formVal.name,
        descriptionTechSpecs: formVal.descriptionTechSpecs
      };

      if (this.isGroupEditMode && this.selectedKey) {
        this.techSpecService.update(this.selectedKey.name, payload).subscribe({
          next: (res: any) => {
            this.isGroupOkLoading = false;
            if (res.status) {
              this.message.success('Group updated');
              this.isGroupModalVisible = false;
              this.loadTechSpecs();
            } else {
              this.message.error(res.message || 'Update failed');
            }
          },
          error: (err) => {
            this.isGroupOkLoading = false;
            this.message.error('Update failed');
          }
        });
      } else {
        this.techSpecService.create(payload).subscribe({
          next: (res: any) => {
            this.isGroupOkLoading = false;
            if (res.status) {
              this.message.success('Group added');
              this.isGroupModalVisible = false;
              this.loadTechSpecs();
            } else {
              this.message.error(res.message || 'Add failed');
            }
          },
          error: (err) => {
            this.isGroupOkLoading = false;
            this.message.error('Add failed');
          }
        });
      }
    } else {
      Object.values(this.groupForm.controls).forEach(c => {
        if (c.invalid) {
          c.markAsDirty();
          c.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  // --- Item actions ---
  onAddItem() {
    if (!this.selectedKey) {
      this.message.warning('Please select a group first');
      return;
    }
    this.isItemEditMode = false;
    this.itemModalTitle = `Add Value to ${this.selectedKey.name}`;
    this.itemForm.reset({ specGroup: this.selectedKey.name });
    this.isItemModalVisible = true;
  }

  onEditItem() {
    if (!this.selectedItem || !this.selectedKey) {
      this.message.warning('Please select an item to edit');
      return;
    }
    this.isItemEditMode = true;
    this.itemModalTitle = 'Edit Tech Spec Value';
    this.itemForm.patchValue({
      id: this.selectedItem.id,
      specGroup: this.selectedKey.name,
      item: this.selectedItem.techSpecs,
      techSpecs: this.selectedItem.item
    });
    this.isItemModalVisible = true;
  }

  onDeleteItem() {
    if (!this.selectedItem || !this.selectedKey) {
      this.message.warning('Please select an item to delete');
      return;
    }
    this.modal.confirm({
      nzTitle: 'Are you sure delete this item?',
      nzContent: `<b style="color: red;">${this.selectedItem.techSpecs}</b>`,
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => {
        this.loadingItems = true;
        this.techSpecService.deleteItem(this.selectedKey!.name, this.selectedItem!.id).subscribe({
          next: (res: any) => {
            if (res.status) {
              this.message.success('Item deleted');
              this.onSelectKey(this.selectedKey!);
            } else {
              this.message.error(res.message || 'Delete failed');
              this.loadingItems = false;
            }
          },
          error: (err) => {
            console.error(err);
            this.message.error('Delete failed');
            this.loadingItems = false;
          }
        });
      }
    });
  }

  handleItemOk() {
    if (this.itemForm.valid && this.selectedKey) {
      this.isItemOkLoading = true;
      const formVal = this.itemForm.getRawValue();
      const payload: TechSpecItem = {
        id: formVal.id,
        item: this.selectedKey.name,
        techSpecs: formVal.item
      };

      if (this.isItemEditMode && this.selectedItem) {
        this.techSpecService.updateItem(this.selectedKey.name, this.selectedItem.id, payload).subscribe({
          next: (res: any) => {
            this.isItemOkLoading = false;
            if (res.status) {
              this.message.success('Item updated');
              this.isItemModalVisible = false;
              this.onSelectKey(this.selectedKey!);
            } else {
              this.message.error(res.message || 'Update failed');
            }
          },
          error: (err) => {
            this.isItemOkLoading = false;
            this.message.error('Update failed');
          }
        });
      } else {
        this.techSpecService.createItem(this.selectedKey.name, payload).subscribe({
          next: (res: any) => {
            this.isItemOkLoading = false;
            if (res.status) {
              this.message.success('Item added');
              this.isItemModalVisible = false;
              this.onSelectKey(this.selectedKey!);
            } else {
              this.message.error(res.message || 'Add failed');
            }
          },
          error: (err) => {
            this.isItemOkLoading = false;
            this.message.error('Add failed');
          }
        });
      }
    } else {
      Object.values(this.itemForm.controls).forEach(c => {
        if (c.invalid) {
          c.markAsDirty();
          c.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }
}
