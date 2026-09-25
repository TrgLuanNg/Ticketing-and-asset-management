import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BaseApiService } from './base-api.service';

export interface PermissionEntity {
  userID: string;
  assetsModify: boolean | null;
  assetsDel: boolean | null;
  departmentModify: boolean | null;
  departmentDel: boolean | null;
  typeModify: boolean | null;
  typeDel: boolean | null;
  userManager: boolean | null;
  disposal: boolean | null;
  permissionApproval: boolean | null;
  allowSeeAllFromAllowedAreas: boolean | null;
}

@Injectable({
  providedIn: 'root'
})
export class PermissionService extends BaseApiService<PermissionEntity, string> {
  protected override get resourceUrl(): string {
    return 'api/permissions';
  }

  constructor(protected override http: HttpClient) {
    super(http);
  }
}
