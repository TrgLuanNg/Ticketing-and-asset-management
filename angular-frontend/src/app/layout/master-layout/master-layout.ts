import { Component, OnInit } from '@angular/core';
import { Router, RouterLink, RouterOutlet } from '@angular/router';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzTooltipModule } from 'ng-zorro-antd/tooltip';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzModalModule, NzModalService } from 'ng-zorro-antd/modal';
import { NzMessageService } from 'ng-zorro-antd/message';
import { CommonModule } from '@angular/common';
import { UserService } from '../../core/services/user.service';

@Component({
  selector: 'app-master-layout',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    RouterLink,
    NzLayoutModule,
    NzMenuModule,
    NzIconModule,
    NzTooltipModule,
    NzButtonModule,
    NzModalModule
  ],
  templateUrl: './master-layout.html',
  styleUrl: './master-layout.less',
  providers: [NzModalService]
})
export class MasterLayout implements OnInit {

  constructor(
    private userService: UserService,
    private router: Router,
    private modal: NzModalService,
    private message: NzMessageService
  ) {}

  ngOnInit(): void {
  }

  isAttributeActive(): boolean {
    const url = this.router.url;
    return url.includes('/department') || 
           url.includes('/type') || 
           url.includes('/tech-specs') || 
           url.includes('/title-employee');
  }

  isAdmin(): boolean {
    const userStr = localStorage.getItem('currentUser');
    if (!userStr) return false;
    try {
      const user = JSON.parse(userStr);
      const roles: string[] = user.roles || user.Roles || [];
      return roles.some(role => typeof role === 'string' && role.toLowerCase() === 'admin');
    } catch {
      return false;
    }
  }

  hasPermission(permission: string): boolean {
    return this.userService.hasPermission(permission);
  }

  getCurrentUserName(): string {
    const userStr = localStorage.getItem('currentUser');
    if (!userStr) return 'User';
    try {
      const user = JSON.parse(userStr);
      return user.userName || user.UserName || user.userId || user.userID || 'User';
    } catch {
      return 'User';
    }
  }

  getCurrentUserInitials(): string {
    const name = this.getCurrentUserName();
    if (!name) return 'U';
    const parts = name.trim().split(' ');
    if (parts.length > 1) {
      return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  }

  getCurrentUserRole(): string {
    const userStr = localStorage.getItem('currentUser');
    if (!userStr) return '';
    try {
      const user = JSON.parse(userStr);
      const roles: string[] = user.roles || user.Roles || [];
      return roles.length > 0 ? roles[0] : 'User';
    } catch {
      return 'User';
    }
  }

  onLogout(): void {
    this.modal.confirm({
      nzTitle: 'Are you sure you want to log out?',
      nzOkText: 'Yes',
      nzOkType: 'primary',
      nzOkDanger: true,
      nzOnOk: () => {
        this.userService.logout().subscribe({
          next: () => {
            localStorage.removeItem('currentUser');
            this.message.success('Logged out successfully');
            this.router.navigate(['/login']);
          },
          error: (err) => {
            console.error('Logout failed', err);
            localStorage.removeItem('currentUser');
            this.router.navigate(['/login']);
          }
        });
      }
    });
  }
}
