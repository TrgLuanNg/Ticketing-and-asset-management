export interface Department {
  id: string;
  name?: string;
  description?: string;
  department?: boolean;
  office?: boolean;
  company?: boolean;
  vendor?: boolean;
  division?: boolean;
  groupDivision?: boolean;
  parentId?: string;
}

export interface AssetType {
  type: string;
  typeName?: string;
  vn?: string;
}

export interface Asset {
  unitCode: string;
  type?: string;
  userName?: string;
  department?: string;
  office?: string;
  hrCode?: string;
  email?: string;
  watt?: string;
  purchasedValue?: string;
  purchasedDate?: string;
  notes?: string;
  model?: string;
  brand?: string;
  status?: string;
  state?: string;
  serial?: string;
  vendor?: string;
  division?: string;
  groupDep?: string;
  depreciation?: number;
  value?: number;
  accessories?: boolean | number;
  ad?: boolean;
  me?: boolean;
  tm?: boolean;
  adOS?: string;
  meTime?: string;
  tmTime?: string;
}

export interface Handover {
  id: string;
  unitCode?: string;
  handoverDate?: Date;
  userCreated?: string;
}

export interface ProblemReport {
  id: string;
  unitCode?: string;
  reportDate?: Date;
  userCreated?: string;
  problemSummary?: string;
}

export interface TechSpecKey {
  name: string;
  descriptionTechSpecs?: string;
}

export interface TechSpecItem {
  id: string;
  item?: string;
  techSpecs?: string;
}

export interface UserInfo {
  userID: string;
  userId?: string;
  userName?: string;
  description?: string;
  area?: string;
  roles?: string[];
  permissions?: string[];
}

export interface TitleEmployee {
  employeeCode: string;
  title?: string;
}

export interface StateMasterData {
  id: string;
  state?: string;
  description?: string;
}
