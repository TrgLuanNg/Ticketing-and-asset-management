import { Routes } from '@angular/router';
import { MasterLayout } from './layout/master-layout/master-layout';
import { inject } from '@angular/core';
import { Router } from '@angular/router';

const authGuard = () => {
  const router = inject(Router);
  if (localStorage.getItem('currentUser')) {
    return true;
  }
  router.navigate(['/login']);
  return false;
};

const adminGuard = () => {
  const router = inject(Router);
  const userStr = localStorage.getItem('currentUser');
  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      const permissions: string[] = user.permissions || user.Permissions || [];
      const roles: string[] = user.roles || user.Roles || [];
      if (permissions.includes('PermissionApproval') || roles.some(role => typeof role === 'string' && role.toLowerCase() === 'admin')) {
        return true;
      }
    } catch {}
  }
  router.navigate(['/assets']);
  return false;
};

const onlyAdminGuard = () => {
  const router = inject(Router);
  const userStr = localStorage.getItem('currentUser');
  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      const roles: string[] = user.roles || user.Roles || [];
      if (roles.some(role => typeof role === 'string' && role.toLowerCase() === 'admin')) {
        return true;
      }
    } catch {}
  }
  router.navigate(['/assets']);
  return false;
};

const userManagerGuard = () => {
  const router = inject(Router);
  const userStr = localStorage.getItem('currentUser');
  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      const roles: string[] = user.roles || user.Roles || [];
      const permissions: string[] = user.permissions || user.Permissions || [];
      if (roles.some(role => typeof role === 'string' && role.toLowerCase() === 'admin') || permissions.includes('UserManager')) {
        return true;
      }
    } catch {}
  }
  router.navigate(['/assets']);
  return false;
};

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./containers/login/login').then(m => m.Login)
  },
  {
    path: 'auth',
    loadComponent: () => import('./containers/authentication/authentication').then(m => m.Authentication)
  },
  {
    path: '',
    component: MasterLayout,
    canActivate: [authGuard],
    children: [
      { path: '', redirectTo: 'statistics', pathMatch: 'full' },
      { path: 'assets', children: [
        { path: '', loadComponent: () => import('./containers/assets/assets').then(m => m.Assets) },
        { path: ':unitCode', loadComponent: () => import('./containers/asset-detail/asset-detail').then(m => m.AssetDetail) },
      ] },
      { path: 'department', loadComponent: () => import('./containers/department/department').then(m => m.Department), canActivate: [onlyAdminGuard] },
      { path: 'type', loadComponent: () => import('./containers/type/type').then(m => m.Type), canActivate: [onlyAdminGuard] },
      { path: 'user', loadComponent: () => import('./containers/user/user').then(m => m.User), canActivate: [userManagerGuard] },
      { path: 'tech-specs', loadComponent: () => import('./containers/tech-specs/tech-specs').then(m => m.TechSpecs), canActivate: [onlyAdminGuard] },
      { path: 'statistics', loadComponent: () => import('./containers/statistics/statistics').then(m => m.Statistics), canActivate: [onlyAdminGuard] },
      { path: 'title-employee', loadComponent: () => import('./containers/title-employee/title-employee').then(m => m.TitleEmployee), canActivate: [onlyAdminGuard] },
    ]
  }
];
