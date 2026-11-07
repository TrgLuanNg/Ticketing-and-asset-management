import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule, FormsModule } from '@angular/forms';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { UserInfo } from '../../core/services/api.service';
import { UserService } from '../../core/services/user.service';
import { UserInfoService, UserInfoEntity } from '../../core/services/user-info.service';
import { PermissionService, PermissionEntity } from '../../core/services/permission.service';

@Component({
  selector: 'app-user',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    FormsModule,
    NzButtonModule,
    NzIconModule,
    NzTableModule,
    NzModalModule,
    NzInputModule,
    NzFormModule,
    NzSelectModule,
    NzTabsModule,
    NzCheckboxModule
  ],
  templateUrl: './user.html',
  styleUrl: './user.less',
  providers: [NzModalService]
})
export class User implements OnInit {
  public allUsers: UserInfo[] = [];
  public users: UserInfo[] = [];
  public loading = false;
  public searchText = '';
  currentFilter: any = {
    pageIndex: 1,
    pageSize: 100
  };

  selectedUser: UserInfo | null = null;
  tabIndex = 0;

  // History Modal
  isHistoryVisible = false;
  historyLoading = false;
  userHistory: any[] = [];

  // Add User Form
  addUserForm!: FormGroup;
  isAddUserModalVisible = false;

  // Edit User Form
  editUserForm!: FormGroup;
  isEditUserModalVisible = false;

  // Edit Role Modal
  isEditRoleModalVisible = false;
  selectedUserRole = 'Viewer';

  // Permission tabs data
  userPermissionsMap: { [userId: string]: PermissionEntity } = {};

  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private userInfoService: UserInfoService,
    private permissionService: PermissionService,
    private cdr: ChangeDetectorRef,
    private message: NzMessageService,
    private modal: NzModalService
  ) { }

  ngOnInit() {
    this.addUserForm = this.fb.group({
      userID: ['', [Validators.required, Validators.maxLength(20)]],
      userName: ['', [Validators.required]]
    });
    this.editUserForm = this.fb.group({
      userID: [{ value: '', disabled: true }, [Validators.required]],
      userName: ['', [Validators.required]]
    });
    this.loadUsers();
    this.loadPermissions();
  }

  loadUsers() {
    this.loading = true;
    this.userService.getAll().subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.allUsers = res.data.map((u: any) => ({
            ...u,
            userID: u.userID || u.userId || u.UserId
          }));
        } else {
          this.allUsers = [];
        }
        this.applyFilter();
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err: any) => {
        console.error('Failed to load users', err);
        this.message.error('Failed to load users');
        this.loading = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  loadPermissions() {
    this.permissionService.getAll().subscribe({
      next: (res: any) => {
        const list = res.data || [];
        this.userPermissionsMap = {};
        list.forEach((p: any) => {
          const uid = p.userID || p.userId || p.UserId;
          if (uid) {
            this.userPermissionsMap[uid] = { ...p, userID: uid };
          }
        });
        this.cdr.detectChanges();
      },
      error: () => {
        this.message.error('Failed to load user permissions.');
        this.cdr.detectChanges();
      }
    });
  }

  applyFilter() {
    let filtered: UserInfo[] = [];
    if (!this.searchText) {
      filtered = [...this.allUsers];
    } else {
      const lower = this.searchText.toLowerCase();
      filtered = this.allUsers.filter(x =>
        (x.userID && x.userID.toLowerCase().includes(lower)) ||
        (x.userName && x.userName.toLowerCase().includes(lower)) ||
        (x.area && x.area.toLowerCase().includes(lower))
      );
    }

    // Sort: Admin -> Viewer -> Inactive
    this.users = filtered.sort((a, b) => {
      const roleA = this.getUserRole(a).toLowerCase();
      const roleB = this.getUserRole(b).toLowerCase();

      const valA = roleA === 'admin' ? 0 : (roleA === 'viewer' ? 1 : 2);
      const valB = roleB === 'admin' ? 0 : (roleB === 'viewer' ? 1 : 2);

      if (valA !== valB) {
        return valA - valB;
      }
      return (a.userID || '').localeCompare(b.userID || '');
    });
  }

  getFilteredUsersForPermissions(): UserInfo[] {
    let list = this.allUsers.filter(u => this.getUserRole(u) !== 'Viewer');
    if (!this.searchText) return list;
    const lower = this.searchText.toLowerCase();
    return list.filter(x =>
      (x.userID && x.userID.toLowerCase().includes(lower)) ||
      (x.userName && x.userName.toLowerCase().includes(lower)) ||
      (x.area && x.area.toLowerCase().includes(lower))
    );
  }

  onSearch(value: string) {
    this.searchText = value;
    this.applyFilter();
  }

  // --- Account Tab Methods ---

  onHistory() {
    if (!this.selectedUser) {
      this.message.warning('Please select a user to view history');
      return;
    }

    this.isHistoryVisible = true;
    this.historyLoading = true;
    this.userHistory = [];

    const uid = this.selectedUser.userID || (this.selectedUser as any).userId;
    this.userService.getHistory(uid).subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          this.userHistory = res.data;
        } else if (Array.isArray(res)) {
          this.userHistory = res;
        }
        this.historyLoading = false;
        setTimeout(() => this.cdr.detectChanges());
      },
      error: (err) => {
        console.error(err);
        this.message.error('Failed to load user history');
        this.historyLoading = false;
        setTimeout(() => this.cdr.detectChanges());
      }
    });
  }

  closeHistory(): void {
    this.isHistoryVisible = false;
  }

  onRowSelect(user: UserInfo): void {
    const currentId = this.selectedUser ? (this.selectedUser.userID || (this.selectedUser as any).userId) : null;
    const targetId = user ? (user.userID || (user as any).userId) : null;
    if (currentId && currentId === targetId) {
      this.selectedUser = null;
    } else {
      this.selectedUser = user;
    }
  }

  hasPermission(permission: string): boolean {
    return this.userService.hasPermission(permission);
  }

  showAddUserModal() {
    this.addUserForm.reset();
    this.isAddUserModalVisible = true;
  }

  handleAddUserOk() {
    if (this.addUserForm.invalid) {
      Object.values(this.addUserForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      return;
    }

    const val = this.addUserForm.value;
    const payload: any = {
      userID: val.userID,
      userName: val.userName,
      pass: 'whatever',
      description: undefined,
      area: undefined,
      roles: ['Viewer'],
      permissions: []
    };

    this.loading = true;
    this.userService.create(payload).subscribe({
      next: (res: any) => {
        this.loading = false;
        if (res.status) {
          this.message.success('User created successfully');
          this.isAddUserModalVisible = false;
          this.loadUsers();
          this.loadPermissions();
        } else {
          this.message.error(res.message || 'Failed to create user');
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.loading = false;
        this.message.error('Failed to create user due to system error');
        this.cdr.detectChanges();
      }
    });
  }

  getUserRole(user: any): string {
    if (user && user.roles && user.roles.length > 0) {
      const r = user.roles[0];
      if (r.toLowerCase() === 'admin') return 'Admin';
      if (r.toLowerCase() === 'inactive') return 'Inactive';
      return 'Viewer';
    }
    return 'Viewer';
  }

  showEditRoleModal() {
    if (!this.selectedUser) {
      this.message.warning('Please select a user first');
      return;
    }
    this.selectedUserRole = this.getUserRole(this.selectedUser);
    this.isEditRoleModalVisible = true;
  }

  handleEditRoleOk() {
    if (!this.selectedUser) return;
    this.loading = true;

    const uid = this.selectedUser.userID || (this.selectedUser as any).userId;
    const updatedUser = {
      ...this.selectedUser,
      roles: [this.selectedUserRole]
    };

    this.userService.update(uid, updatedUser).subscribe({
      next: (res: any) => {
        this.loading = false;
        if (res.status) {
          this.message.success('User role updated successfully');
          this.isEditRoleModalVisible = false;
          this.loadUsers();
          this.selectedUser = null;
        } else {
          this.message.error(res.message || 'Failed to update user role');
        }
        this.cdr.detectChanges();
      }
    });
  }

  showEditUserModal() {
    if (!this.selectedUser) {
      this.message.warning('Please select a user first');
      return;
    }
    const uid = this.selectedUser.userID || (this.selectedUser as any).userId;
    this.editUserForm.patchValue({
      userID: uid,
      userName: this.selectedUser.userName
    });
    this.isEditUserModalVisible = true;
  }

  handleEditUserOk() {
    if (this.editUserForm.invalid) {
      Object.values(this.editUserForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
      return;
    }

    if (!this.selectedUser) return;
    this.loading = true;

    const uid = this.selectedUser.userID || (this.selectedUser as any).userId;
    const val = this.editUserForm.getRawValue();
    const updatedUser = {
      ...this.selectedUser,
      userName: val.userName
    };

    this.userService.update(uid, updatedUser).subscribe({
      next: (res: any) => {
        this.loading = false;
        if (res.status) {
          this.message.success('User updated successfully');
          this.isEditUserModalVisible = false;
          this.loadUsers();
          this.selectedUser = null;
        } else {
          this.message.error(res.message || 'Failed to update user');
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.loading = false;
        console.error(err);
        this.message.error('An error occurred while updating user');
        this.cdr.detectChanges();
      }
    });
  }

  // --- Permission Tab Methods ---

  getOrCreatePermission(userId: string): PermissionEntity {
    if (!this.userPermissionsMap[userId]) {
      this.userPermissionsMap[userId] = {
        userID: userId,
        assetsModify: null,
        assetsDel: null,
        departmentModify: null,
        departmentDel: null,
        typeModify: null,
        typeDel: null,
        userManager: null,
        disposal: null,
        permissionApproval: null,
        allowSeeAllFromAllowedAreas: null
      };
    }
    return this.userPermissionsMap[userId];
  }

  hasArea(user: UserInfo, areaName: string): boolean {
    if (!user.area) return false;
    const areas = user.area.split(',').map(a => a.trim().toLowerCase());
    return areas.includes(areaName.toLowerCase());
  }

  toggleArea(user: UserInfo, areaName: string, checked: boolean) {
    let currentAreas = user.area ? user.area.split(',').map(a => a.trim()).filter(a => a.length > 0) : [];

    if (checked) {
      if (!currentAreas.includes(areaName)) {
        currentAreas.push(areaName);
      }
    } else {
      currentAreas = currentAreas.filter(a => a.toLowerCase() !== areaName.toLowerCase());
    }

    user.area = currentAreas.join(', ');
  }

  saveUserPermission(user: UserInfo) {
    this.userInfoService.update(user.userID, user).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success(`Allowed areas updated for user ${user.userID}`);
        } else {
          this.message.error(res.message || 'Failed to update user areas.');
        }
      },
      error: () => {
        this.message.error('Failed to update user areas.');
      }
    });
  }

  resetPermissions(userId: string) {
    this.userPermissionsMap[userId] = {
      userID: userId,
      assetsModify: null,
      assetsDel: null,
      departmentModify: null,
      departmentDel: null,
      typeModify: null,
      typeDel: null,
      userManager: null,
      disposal: null,
      permissionApproval: null,
      allowSeeAllFromAllowedAreas: null
    };
    this.cdr.detectChanges();
  }

  saveUserPermissionControl(userId: string) {
    const perm = this.getOrCreatePermission(userId);
    this.permissionService.update(userId, perm).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success(`Permissions updated for user ${userId}`);

          const userStr = localStorage.getItem('currentUser');
          if (userStr) {
            const user = JSON.parse(userStr);
            if (user.userId === userId || user.userID === userId) {
              const newPerms: string[] = [];
              if (perm.assetsModify === true) newPerms.push('AssetsModify');
              if (perm.assetsDel === true) newPerms.push('AssetsDel');
              if (perm.departmentModify === true) newPerms.push('DepartmentModify');
              if (perm.departmentDel === true) newPerms.push('DepartmentDel');
              if (perm.typeModify === true) newPerms.push('TypeModify');
              if (perm.typeDel === true) newPerms.push('TypeDel');
              if (perm.userManager === true) newPerms.push('UserManager');
              if (perm.disposal === true) newPerms.push('Disposal');
              if (perm.permissionApproval === true) newPerms.push('PermissionApproval');
              if (perm.allowSeeAllFromAllowedAreas === true) newPerms.push('AllowSeeAllFromAllowedAreas');

              user.permissions = newPerms;
              user.Permissions = newPerms;
              localStorage.setItem('currentUser', JSON.stringify(user));
            }
          }
        } else {
          this.message.error(res.message || 'Failed to update permissions.');
        }
      },
      error: () => {
        this.message.error('Failed to update permissions.');
      }
    });
  }
}
