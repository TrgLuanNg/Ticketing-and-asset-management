import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTabsModule } from 'ng-zorro-antd/tabs';
import { NzInputModule } from 'ng-zorro-antd/input';
import { UserInfoService, UserInfoEntity } from '../../core/services/user-info.service';
import { PermissionService, PermissionEntity } from '../../core/services/permission.service';

@Component({
  selector: 'app-permission',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    NzTableModule,
    NzButtonModule,
    NzCheckboxModule,
    NzIconModule,
    NzTabsModule,
    NzInputModule
  ],
  templateUrl: './permission.html',
  styleUrls: ['./permission.less']
})
export class Permission implements OnInit {
  users: UserInfoEntity[] = [];
  userPermissionsMap: { [userId: string]: PermissionEntity } = {};
  loading = false;
  searchText = '';

  getFilteredUsers(): UserInfoEntity[] {
    if (!this.searchText) {
      return this.users;
    }
    const query = this.searchText.trim().toLowerCase();
    return this.users.filter(u => 
      (u.userID && u.userID.toLowerCase().includes(query)) || 
      (u.userName && u.userName.toLowerCase().includes(query))
    );
  }

  constructor(
    private userInfoService: UserInfoService,
    private permissionService: PermissionService,
    private message: NzMessageService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    this.loadData();
  }

  loadData() {
    this.loading = true;
    this.userInfoService.search({ pageIndex: 1, pageSize: 1000 }).subscribe({
      next: (res: any) => {
        this.users = res.data || [];
        this.loadPermissions();
      },
      error: () => {
        this.message.error('Failed to load users.');
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  loadPermissions() {
    this.permissionService.getAll().subscribe({
      next: (res: any) => {
        const list = res.data || [];
        this.userPermissionsMap = {};
        list.forEach((p: PermissionEntity) => {
          this.userPermissionsMap[p.userID] = p;
        });
        this.loading = false;
        this.cdr.detectChanges();
      },
      error: () => {
        this.message.error('Failed to load user permissions.');
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  getOrCreatePermission(userId: string): PermissionEntity {
    if (!this.userPermissionsMap[userId]) {
      this.userPermissionsMap[userId] = {
        userID: userId,
        assetsModify: false,
        assetsDel: false,
        departmentModify: false,
        departmentDel: false,
        typeModify: false,
        typeDel: false,
        userManager: false,
        disposal: false,
        permissionApproval: false,
        allowSeeAllFromAllowedAreas: false
      };
    }
    return this.userPermissionsMap[userId];
  }

  hasArea(user: UserInfoEntity, areaName: string): boolean {
    if (!user.area) return false;
    const areas = user.area.split(',').map(a => a.trim().toLowerCase());
    return areas.includes(areaName.toLowerCase());
  }

  toggleArea(user: UserInfoEntity, areaName: string, checked: boolean) {
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

  saveUserPermission(user: UserInfoEntity) {
    this.userInfoService.update(user.userID, user).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success(`Allowed areas updated for user ${user.userID}`);
          
          // Sync localStorage if editing own areas
          const userStr = localStorage.getItem('currentUser');
          if (userStr) {
            const currentUser = JSON.parse(userStr);
            const currentUserId = currentUser.userId || currentUser.userID || currentUser.UserId;
            if (currentUserId === user.userID) {
              currentUser.area = user.area;
              currentUser.Area = user.area;
              localStorage.setItem('currentUser', JSON.stringify(currentUser));
            }
          }
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
      assetsModify: false,
      assetsDel: false,
      departmentModify: false,
      departmentDel: false,
      typeModify: false,
      typeDel: false,
      userManager: false,
      disposal: false,
      permissionApproval: false,
      allowSeeAllFromAllowedAreas: false
    };
    this.cdr.detectChanges();
  }

  saveUserPermissionControl(userId: string) {
    const perm = this.getOrCreatePermission(userId);
    this.permissionService.update(userId, perm).subscribe({
      next: (res: any) => {
        if (res.status) {
          this.message.success(`Permissions updated for user ${userId}`);
          
          // Sync localStorage if editing own permissions
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
