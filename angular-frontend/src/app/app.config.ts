import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter } from '@angular/router';
import { IMAGE_CONFIG } from '@angular/common';
import { routes } from './app.routes';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideNzIcons } from 'ng-zorro-antd/icon';
import { en_US, provideNzI18n } from 'ng-zorro-antd/i18n';
import { RouteReuseStrategy } from '@angular/router';
import { GridRouteReuseStrategy } from './grid-route-reuse.strategy';
import { authInterceptor } from './core/interceptors/auth.interceptor';
import { NzMessageService } from 'ng-zorro-antd/message';
import { CustomMessageService } from './core/services/custom-message.service';
import {
  MenuFoldOutline,
  MenuUnfoldOutline,
  DashboardOutline,
  AppstoreOutline,
  SettingOutline,
  FilterOutline,
  DownloadOutline,
  ReloadOutline,
  PlusOutline,
  FilePdfOutline,
  LogoutOutline,
  ApartmentOutline,
  TagOutline,
  IdcardOutline,
  UserOutline,
  LockOutline,
  HistoryOutline,
  DeleteOutline,
  UploadOutline,
  PieChartOutline,
  SwapOutline,
  FileTextOutline,
  InfoCircleOutline,
  EditOutline,
  ArrowLeftOutline,
  BookOutline,
  TableOutline,
  BarChartOutline,
  PrinterOutline
} from '@ant-design/icons-angular/icons';

const icons = [
  MenuFoldOutline,
  MenuUnfoldOutline,
  DashboardOutline,
  AppstoreOutline,
  SettingOutline,
  FilterOutline,
  DownloadOutline,
  ReloadOutline,
  PlusOutline,
  FilePdfOutline,
  LogoutOutline,
  ApartmentOutline,
  TagOutline,
  IdcardOutline,
  UserOutline,
  LockOutline,
  HistoryOutline,
  DeleteOutline,
  UploadOutline,
  PieChartOutline,
  SwapOutline,
  FileTextOutline,
  InfoCircleOutline,
  EditOutline,
  ArrowLeftOutline,
  BookOutline,
  TableOutline,
  BarChartOutline,
  PrinterOutline
];

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideAnimationsAsync(),
    provideHttpClient(withInterceptors([authInterceptor])),
    provideNzIcons(icons),
    provideNzI18n(en_US),
    { provide: NzMessageService, useClass: CustomMessageService },
    { provide: RouteReuseStrategy, useClass: GridRouteReuseStrategy },
    {
      provide: IMAGE_CONFIG,
      useValue: {
        disableImageSizeWarning: true, 
        disableImageLazyLoadWarning: true
      }
    }
  ]
};