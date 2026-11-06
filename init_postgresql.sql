CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE SCHEMA IF NOT EXISTS public;
SET search_path TO public;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'datetime') THEN
        CREATE DOMAIN datetime AS timestamp with time zone;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'image') THEN
        CREATE DOMAIN image AS bytea;
    END IF;
END $$;

DROP TABLE IF EXISTS "ACCESSORIES", accessories CASCADE;
DROP TABLE IF EXISTS "ASSETSLOGS", assets_logs, assetslogs CASCADE;
DROP TABLE IF EXISTS "DISPOSAL", disposal CASCADE;
DROP TABLE IF EXISTS "HANDOVER", handover CASCADE;
DROP TABLE IF EXISTS "HANDOVEROTHERS", handover_others, handoverothers CASCADE;
DROP TABLE IF EXISTS "PROBLEMREPORT", problem_report, problemreport CASCADE;
DROP TABLE IF EXISTS "SERVICE", service CASCADE;
DROP TABLE IF EXISTS "SOFTWARE", software CASCADE;
DROP TABLE IF EXISTS "UpdateEmployeeCode", update_employee_code CASCADE;
DROP TABLE IF EXISTS "ASSETS", assets CASCADE;
DROP TABLE IF EXISTS "ASSETS_bak20240711", assets_bak20240711 CASCADE;
DROP TABLE IF EXISTS "ASSETSAD", assets_ad, assetsad CASCADE;
DROP TABLE IF EXISTS "ASSETSCHEKING", assets_checking, assetscheking CASCADE;
DROP TABLE IF EXISTS "ASSETSEMPATCHDETAIL", assets_em_patch_detail, assetsempatchdetail CASCADE;
DROP TABLE IF EXISTS "ASSETSME", assets_me, assetsme CASCADE;
DROP TABLE IF EXISTS "ASSETSTM", assets_tm, assetstm CASCADE;
DROP TABLE IF EXISTS "CheckingResult", checking_result CASCADE;
DROP TABLE IF EXISTS "COMBO", combo CASCADE;
DROP TABLE IF EXISTS "DEPARTMENT", department CASCADE;
DROP TABLE IF EXISTS "DEVICECONFIG", device_config, deviceconfig CASCADE;
DROP TABLE IF EXISTS "Documents", documents CASCADE;
DROP TABLE IF EXISTS "EmployeeTitle", employee_title CASCADE;
DROP TABLE IF EXISTS "HISTORY", history CASCADE;
DROP TABLE IF EXISTS "IMAGEREPORT", image_report, imagereport CASCADE;
DROP TABLE IF EXISTS "ITEMMASTERDATA", item_master_data, itemmasterdata CASCADE;
DROP TABLE IF EXISTS "KEYTECHSPECS", key_tech_specs, keytechspecs CASCADE;
DROP TABLE IF EXISTS "PERMISSION", permission CASCADE;
DROP TABLE IF EXISTS "PREFIXUNITCODE", prefix_unit_code, prefixunitcode CASCADE;
DROP TABLE IF EXISTS "REFRESH_TOKENS", refresh_tokens CASCADE;
DROP TABLE IF EXISTS "STATEMASTERDATA", state_master_data, statemasterdata CASCADE;
DROP TABLE IF EXISTS "TYPE", type CASCADE;
DROP TABLE IF EXISTS "USERINFO", user_info, userinfo CASCADE;
DROP TABLE IF EXISTS "test", test CASCADE;

CREATE TABLE department (
    id VARCHAR(5) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description VARCHAR(250),
    department BOOLEAN NOT NULL DEFAULT FALSE,
    office BOOLEAN NOT NULL DEFAULT FALSE,
    company BOOLEAN NOT NULL DEFAULT FALSE,
    vendor BOOLEAN NOT NULL DEFAULT FALSE,
    division BOOLEAN DEFAULT FALSE,
    group_division BOOLEAN DEFAULT FALSE,
    parent_id VARCHAR(50)
);

CREATE INDEX idx_department_id_name ON department (id, name);

-- -----------------------------------------------------------------------------
-- Table: type (TYPE)
-- -----------------------------------------------------------------------------
CREATE TABLE type (
    type VARCHAR(20) PRIMARY KEY,
    type_name VARCHAR(50) NOT NULL,
    vn VARCHAR(50)
);

CREATE TABLE state_master_data (
    id VARCHAR(3) PRIMARY KEY,
    state VARCHAR(50) NOT NULL,
    description VARCHAR(50)
);

CREATE TABLE user_info (
    user_id VARCHAR(20) PRIMARY KEY,
    pass VARCHAR(255) NOT NULL,
    user_name VARCHAR(50) NOT NULL,
    description VARCHAR(150),
    area VARCHAR(40),
    role VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    failed_login_attempts INTEGER DEFAULT 0,
    lockout_end TIMESTAMPTZ
);

CREATE TABLE permission (
    user_id VARCHAR(20) PRIMARY KEY REFERENCES user_info(user_id) ON DELETE CASCADE,
    assets_modify BOOLEAN DEFAULT FALSE,
    assets_del BOOLEAN DEFAULT FALSE,
    department_modify BOOLEAN DEFAULT FALSE,
    department_del BOOLEAN DEFAULT FALSE,
    type_modify BOOLEAN DEFAULT FALSE,
    type_del BOOLEAN DEFAULT FALSE,
    user_manager BOOLEAN DEFAULT FALSE,
    disposal BOOLEAN DEFAULT FALSE,
    permission_approval BOOLEAN DEFAULT FALSE,
    allow_see_all_from_allowed_areas BOOLEAN DEFAULT FALSE
);

-- -----------------------------------------------------------------------------
-- Table: refresh_tokens (REFRESH_TOKENS)
-- -----------------------------------------------------------------------------
CREATE TABLE refresh_tokens (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    token VARCHAR(200) NOT NULL,
    user_id VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ,
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    replaced_by_token VARCHAR(200)
);

CREATE INDEX idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE TABLE assets (
    unit_code VARCHAR(21) PRIMARY KEY,
    type VARCHAR(3) NOT NULL,
    user_name VARCHAR(50),
    detail1 VARCHAR(6),
    detail2 VARCHAR(6),
    detail3 VARCHAR(6),
    detail4 VARCHAR(6),
    detail5 VARCHAR(6),
    detail6 VARCHAR(6),
    detail7 VARCHAR(6),
    detail8 VARCHAR(6),
    detail9 VARCHAR(6),
    model VARCHAR(50),
    id_tag VARCHAR(50),
    purchased_date DATE,
    usage_date DATE,
    brand VARCHAR(20),
    warranty_expiration INTEGER,
    purchased_value NUMERIC(18, 0),
    vendor VARCHAR(5),
    type_exp VARCHAR(10),
    notes TEXT,
    department VARCHAR(5),
    office VARCHAR(5),
    area VARCHAR(5),
    company VARCHAR(20),
    status VARCHAR(20),
    serial VARCHAR(50),
    record_date DATE,
    hr_code VARCHAR(10),
    depreciation NUMERIC(18, 0),
    value NUMERIC(18, 0),
    accessories BOOLEAN DEFAULT FALSE,
    main_asset VARCHAR(15),
    email VARCHAR(50),
    state VARCHAR(3),
    watt VARCHAR(50),
    group_dep VARCHAR(5),
    division VARCHAR(5)
);

CREATE INDEX idx_assets_details ON assets (
    detail1, detail2, detail3, detail4, detail5, detail6, detail7, detail8, detail9
);

CREATE INDEX idx_assets_grid_covering ON assets (
    area ASC,
    purchased_date DESC NULLS LAST,
    unit_code DESC
)
INCLUDE (
    type, user_name, department, office, company,
    division, group_dep, state, status, hr_code,
    notes, email, watt, purchased_value,
    brand, model, serial, vendor,
    detail1, detail2, detail3, detail4, detail5,
    detail6, detail7, detail8, detail9
)
WHERE LOWER(type) IN ('dpc', 'ltp')
  AND state IN ('001', '002', '003');

CREATE INDEX idx_assets_department ON assets (department);
CREATE INDEX idx_assets_state ON assets (state);
CREATE INDEX idx_assets_area ON assets (area);
CREATE INDEX idx_assets_user_name ON assets (user_name);

CREATE TABLE accessories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    unit_code VARCHAR(21) NOT NULL REFERENCES assets(unit_code) ON DELETE CASCADE,
    brand VARCHAR(20),
    description VARCHAR(150),
    serial VARCHAR(50)
);

CREATE INDEX idx_accessories_unit_code ON accessories (unit_code);

CREATE TABLE assets_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    unit_code VARCHAR(21) NOT NULL,
    user_id VARCHAR(20) NOT NULL,
    action VARCHAR(50) NOT NULL,
    action_logs TEXT NOT NULL,
    date_modified TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_assets_logs_unit_code ON assets_logs (unit_code);
CREATE INDEX idx_assets_logs_date_modified ON assets_logs (date_modified);

CREATE TABLE device_config (
    device VARCHAR(3) PRIMARY KEY,
    detail1 VARCHAR(50),
    detail2 VARCHAR(50),
    detail3 VARCHAR(50),
    detail4 VARCHAR(50),
    detail5 VARCHAR(50),
    detail6 VARCHAR(50),
    detail7 VARCHAR(50),
    detail8 VARCHAR(50),
    detail9 VARCHAR(50)
);

CREATE TABLE disposal (
    unit_code VARCHAR(21) NOT NULL,
    id_disposal VARCHAR(8) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status BOOLEAN NOT NULL DEFAULT FALSE,
    note VARCHAR(300),
    user_id VARCHAR(20),
    PRIMARY KEY (unit_code, id_disposal)
);

CREATE INDEX idx_disposal_unit_code ON disposal (unit_code);
CREATE INDEX idx_disposal_id_disposal ON disposal (id_disposal);

CREATE TABLE documents (
    document_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    reference_id VARCHAR(50),
    user_stamp VARCHAR(20),
    date_time_stamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    file_name VARCHAR(100),
    file_extension VARCHAR(10),
    file_description VARCHAR(150),
    file_binary BYTEA
);

CREATE INDEX idx_documents_reference_id ON documents (reference_id);

CREATE TABLE employee_title (
    employee_code VARCHAR(50) PRIMARY KEY,
    title VARCHAR(250)
);

CREATE TABLE handover (
    id VARCHAR(9) NOT NULL,
    unit_code VARCHAR(21) NOT NULL,
    handover_date DATE NOT NULL,
    user_created VARCHAR(20) NOT NULL,
    form VARCHAR(20) NOT NULL,
    user_transfer VARCHAR(50),
    department_transfer VARCHAR(5),
    office_transfer VARCHAR(5),
    company_transfer VARCHAR(5),
    area_transfer VARCHAR(5),
    type VARCHAR(3),
    jira_code VARCHAR(25),
    detail1 VARCHAR(6),
    detail2 VARCHAR(6),
    detail3 VARCHAR(6),
    detail4 VARCHAR(6),
    detail5 VARCHAR(6),
    detail6 VARCHAR(6),
    detail7 VARCHAR(6),
    detail8 VARCHAR(6),
    detail9 VARCHAR(6),
    PRIMARY KEY (id, unit_code)
);

CREATE INDEX idx_handover_id ON handover (id);
CREATE INDEX idx_handover_unit_code ON handover (unit_code);
CREATE INDEX idx_handover_date ON handover (handover_date);

CREATE TABLE handover_others (
    id VARCHAR(9) NOT NULL,
    unit_code VARCHAR(21) NOT NULL,
    handover_date DATE NOT NULL,
    user_created VARCHAR(20) NOT NULL,
    form VARCHAR(20) NOT NULL,
    user_a VARCHAR(50),
    department_a VARCHAR(5),
    office_a VARCHAR(5),
    company_a VARCHAR(5),
    area_a VARCHAR(5),
    user_b VARCHAR(50),
    department_b VARCHAR(5),
    office_b VARCHAR(5),
    company_b VARCHAR(5),
    area_b VARCHAR(5),
    type VARCHAR(3),
    detail1 VARCHAR(6),
    detail2 VARCHAR(6),
    detail3 VARCHAR(6),
    detail4 VARCHAR(6),
    detail5 VARCHAR(6),
    detail6 VARCHAR(6),
    detail7 VARCHAR(6),
    detail8 VARCHAR(6),
    detail9 VARCHAR(6),
    state_a VARCHAR(3),
    state_b VARCHAR(3),
    status_a VARCHAR(10),
    status_b VARCHAR(10),
    handover_no VARCHAR(15),
    jira_code VARCHAR(25),
    PRIMARY KEY (id, unit_code)
);

CREATE INDEX idx_handover_others_id ON handover_others (id);
CREATE INDEX idx_handover_others_unit_code ON handover_others (unit_code);

CREATE TABLE history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL,
    action VARCHAR(250) NOT NULL,
    date_time_ac TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_history_user_id ON history (user_id);
CREATE INDEX idx_history_date_time_ac ON history (date_time_ac);

CREATE TABLE image_report (
    id VARCHAR(9) NOT NULL,
    type_report VARCHAR(20) NOT NULL,
    image1 BYTEA,
    image2 BYTEA,
    PRIMARY KEY (id, type_report)
);

CREATE TABLE item_master_data (
    id VARCHAR(6) PRIMARY KEY,
    item VARCHAR(20) NOT NULL,
    tech_specs VARCHAR(50) NOT NULL
);

CREATE INDEX idx_item_master_data_id ON item_master_data (id);

CREATE TABLE key_tech_specs (
    name VARCHAR(20) PRIMARY KEY,
    description_tech_specs VARCHAR(50)
);

CREATE TABLE prefix_unit_code (
    id_department VARCHAR(5) PRIMARY KEY,
    prefix VARCHAR(3),
    area VARCHAR(5)
);

CREATE TABLE problem_report (
    id VARCHAR(9) NOT NULL,
    unit_code VARCHAR(21) NOT NULL,
    report_date DATE NOT NULL,
    user_created VARCHAR(20) NOT NULL,
    problem_summary VARCHAR(350),
    test_procedures VARCHAR(250),
    solutions VARCHAR(450),
    type VARCHAR(3),
    detail1 VARCHAR(6),
    detail2 VARCHAR(6),
    detail3 VARCHAR(6),
    detail4 VARCHAR(6),
    detail5 VARCHAR(6),
    detail6 VARCHAR(6),
    detail7 VARCHAR(6),
    detail8 VARCHAR(6),
    detail9 VARCHAR(6),
    user_report VARCHAR(50),
    department_report VARCHAR(5),
    office_report VARCHAR(5),
    company_report VARCHAR(5),
    area_report VARCHAR(5),
    problem_no VARCHAR(18),
    jira_code VARCHAR(25),
    PRIMARY KEY (id, unit_code)
);

CREATE INDEX idx_problem_report_id ON problem_report (id);
CREATE INDEX idx_problem_report_unit_code ON problem_report (unit_code);
CREATE INDEX idx_problem_report_date ON problem_report (report_date);

CREATE TABLE service (
    main_assets VARCHAR(15) NOT NULL,
    sub_assets VARCHAR(15) NOT NULL,
    service_level VARCHAR(20) NOT NULL,
    time_create_tag DATE,
    PRIMARY KEY (main_assets, sub_assets)
);

-- -----------------------------------------------------------------------------
-- Table: software (SOFTWARE)
-- -----------------------------------------------------------------------------
CREATE TABLE software (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    unit_code VARCHAR(15) NOT NULL,
    software VARCHAR(50) NOT NULL,
    version VARCHAR(20)
);

CREATE INDEX idx_software_unit_code ON software (unit_code);

CREATE TABLE update_employee_code (
    unit_code VARCHAR(21) PRIMARY KEY,
    hr_code VARCHAR(10)
);

CREATE TABLE combo (
    combo_name VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL,
    brand VARCHAR(20),
    detail1 VARCHAR(6),
    detail2 VARCHAR(6),
    detail3 VARCHAR(6),
    detail4 VARCHAR(6),
    detail5 VARCHAR(6),
    detail6 VARCHAR(6),
    detail7 VARCHAR(6),
    detail8 VARCHAR(6),
    detail9 VARCHAR(6),
    model VARCHAR(50),
    purchased_value NUMERIC(18, 0),
    type_exp VARCHAR(5),
    warranty_expiration INTEGER,
    PRIMARY KEY (combo_name, type)
);

CREATE TABLE checking_result (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    checking_result VARCHAR(250) NOT NULL
);

CREATE TABLE test (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    value_action VARCHAR(500),
    date_action TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets_ad (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    operating_system VARCHAR(150),
    service_pack VARCHAR(150),
    version VARCHAR(50),
    create_year VARCHAR(50),
    cn VARCHAR(50),
    last_logon VARCHAR(50),
    status VARCHAR(50)
);

CREATE INDEX idx_assets_ad_cn ON assets_ad (cn);

CREATE TABLE assets_checking (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    checking_pc VARCHAR(50),
    checking_result VARCHAR(250),
    checking_action VARCHAR(50),
    checking_remark VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets_em_patch_detail (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sqnumber VARCHAR(300),
    os_language VARCHAR(300),
    install_error_code VARCHAR(300),
    bulletin_id VARCHAR(300),
    update_name VARCHAR(300),
    patch_size VARCHAR(300),
    resource_id_string VARCHAR(300),
    uninstall_for VARCHAR(300),
    patch_uninstall_status VARCHAR(300),
    service_pack VARCHAR(300),
    patch_status VARCHAR(300),
    patch_install_status_image VARCHAR(300),
    patch_id VARCHAR(300),
    patch_name VARCHAR(300),
    patch_description VARCHAR(300),
    vendor_name VARCHAR(300),
    superceded_by VARCHAR(300),
    branch_office_name VARCHAR(300),
    platform_name VARCHAR(300),
    severity VARCHAR(300),
    patch_released_time VARCHAR(300),
    resource_id VARCHAR(300),
    os_name VARCHAR(300),
    customer_name VARCHAR(300),
    resource_name VARCHAR(300),
    user_id VARCHAR(300),
    managedcomputer_resource_id VARCHAR(300),
    first_name VARCHAR(300),
    friendly_name VARCHAR(300),
    i18n VARCHAR(300),
    resourcemacip_resource_id VARCHAR(300),
    patch_updated_time VARCHAR(50),
    patch_affected_status_label VARCHAR(50)
);


CREATE TABLE assets_me (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    modernappdetails_remarks VARCHAR(150),
    processor_arch_name VARCHAR(150),
    previous_installation_status VARCHAR(150),
    remarks_args VARCHAR(150),
    agent_installed_on VARCHAR(150),
    agent_executed_on VARCHAR(150),
    os_platform_name VARCHAR(150),
    resource_id_string VARCHAR(150),
    service_pack VARCHAR(150),
    processor_arch_type_id VARCHAR(150),
    processor_architecture VARCHAR(150),
    os_platform VARCHAR(150),
    agent_logged_on_users VARCHAR(3050),
    osflavor_id VARCHAR(150),
    search_tag VARCHAR(150),
    status_label VARCHAR(150),
    error_kb_url VARCHAR(150),
    full_name VARCHAR(150),
    owner_email_id VARCHAR(150),
    operation_type VARCHAR(150),
    agent_upgraded_on VARCHAR(150),
    installation_status VARCHAR(150),
    error_code VARCHAR(150),
    computer_status_update_time VARCHAR(150),
    managed_status VARCHAR(150),
    branch_office_name VARCHAR(150),
    agent_last_contact_time VARCHAR(150),
    description VARCHAR(150),
    domain_netbios_name VARCHAR(150),
    fqdn_name VARCHAR(150),
    mac_address VARCHAR(150),
    owner VARCHAR(150),
    os_version VARCHAR(150),
    last_sync_time VARCHAR(150),
    ip_address VARCHAR(150),
    agent_uninstalled_on VARCHAR(150),
    agent_version VARCHAR(150),
    agent_last_bootup_time VARCHAR(150),
    resource_id VARCHAR(150),
    os_name VARCHAR(150),
    location VARCHAR(150),
    computer_live_status VARCHAR(150),
    customer_name VARCHAR(150),
    resource_name VARCHAR(150),
    customer_id VARCHAR(150),
    remarks VARCHAR(150)
);

CREATE INDEX idx_assets_me_resource_name ON assets_me (resource_name);

CREATE TABLE assets_tm (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    endpoint_id VARCHAR(50),
    endpoint_host VARCHAR(250),
    endpoint_ip VARCHAR(50),
    endpoint_mac VARCHAR(250),
    product VARCHAR(350),
    managing_server_id VARCHAR(50),
    ad_domain VARCHAR(150),
    domain VARCHAR(150),
    domain_hierarchy VARCHAR(50),
    logon_user VARCHAR(150),
    platform VARCHAR(150),
    client_program VARCHAR(50),
    connection_status VARCHAR(30),
    isolation_status VARCHAR(50),
    firewall VARCHAR(20),
    scan_method VARCHAR(50),
    update_agent VARCHAR(20),
    last_scheduled_scan_utc TIMESTAMPTZ,
    last_manual_scan_utc TIMESTAMPTZ,
    last_startup TIMESTAMPTZ,
    last_connected TIMESTAMPTZ,
    virus_scan_engine VARCHAR(30),
    virus_pattern VARCHAR(20),
    smart_scan_agent_pattern VARCHAR(250),
    capabilities VARCHAR(500),
    last_contact_time_category VARCHAR(20)
);

CREATE INDEX idx_assets_tm_endpoint_host ON assets_tm (endpoint_host);

CREATE TABLE assets_bak20240711 (LIKE assets INCLUDING ALL);

CREATE OR REPLACE FUNCTION fn_audit_assets_update()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO assets_logs (unit_code, user_id, action, action_logs, date_modified)
    VALUES (
        NEW.unit_code,
        COALESCE(current_setting('app.current_user_id', true), NEW.user_name, 'SYSTEM'),
        'UPDATE',
        CONCAT('Updated asset: ', NEW.unit_code, ' - State: ', NEW.state, ', Department: ', NEW.department),
        CURRENT_TIMESTAMP
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_assets_update_log ON assets;
CREATE TRIGGER trg_assets_update_log
AFTER UPDATE ON assets
FOR EACH ROW
EXECUTE FUNCTION fn_audit_assets_update();

CREATE OR REPLACE VIEW "DEPARTMENT" AS
SELECT
    id AS "ID",
    name AS "Name",
    description AS "Description",
    department AS "Department",
    office AS "Office",
    company AS "Company",
    vendor AS "Vendor",
    division AS "Division",
    group_division AS "GroupDivision",
    parent_id AS "ParentId"
FROM department;

CREATE OR REPLACE VIEW "TYPE" AS
SELECT
    type AS "Type",
    type_name AS "TypeName",
    vn AS "VN"
FROM type;

CREATE OR REPLACE VIEW "STATEMASTERDATA" AS
SELECT
    id AS "ID",
    state AS "State",
    description AS "Description"
FROM state_master_data;

CREATE OR REPLACE VIEW "USERINFO" AS
SELECT
    user_id AS "UserID",
    pass AS "Pass",
    user_name AS "UserName",
    description AS "Description",
    area AS "Area",
    role AS "Role",
    is_active AS "IsActive",
    failed_login_attempts AS "FailedLoginAttempts",
    lockout_end AS "LockoutEnd"
FROM user_info;

CREATE OR REPLACE VIEW "PERMISSION" AS
SELECT
    user_id AS "UserID",
    assets_modify AS "AssetsModify",
    assets_del AS "AssetsDel",
    department_modify AS "DepartmentModify",
    department_del AS "DepartmentDel",
    type_modify AS "TypeModify",
    type_del AS "TypeDel",
    user_manager AS "UserManager",
    disposal AS "Disposal",
    permission_approval AS "PermissionApproval",
    allow_see_all_from_allowed_areas AS "AllowSeeAllFromAllowedAreas"
FROM permission;

CREATE OR REPLACE VIEW "REFRESH_TOKENS" AS
SELECT
    id AS "Id",
    token AS "Token",
    user_id AS "UserId",
    expires_at AS "ExpiresAt",
    is_revoked AS "IsRevoked",
    created_at AS "CreatedAt",
    replaced_by_token AS "ReplacedByToken"
FROM refresh_tokens;

CREATE OR REPLACE VIEW "ASSETS" AS
SELECT
    unit_code AS "UnitCode",
    type AS "Type",
    user_name AS "UserName",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9",
    model AS "Model",
    id_tag AS "IDTag",
    purchased_date AS "PurchasedDate",
    usage_date AS "UsageDate",
    brand AS "Brand",
    warranty_expiration AS "WarrantyExpiration",
    purchased_value AS "PurchasedValue",
    vendor AS "Vendor",
    type_exp AS "TypeEXP",
    notes AS "Notes",
    department AS "Department",
    office AS "Office",
    area AS "Area",
    company AS "Company",
    status AS "Status",
    serial AS "Serial",
    record_date AS "RecordDate",
    hr_code AS "HRCode",
    depreciation AS "Depreciation",
    value AS "Value",
    accessories AS "Accessories",
    main_asset AS "MainAsset",
    email AS "Email",
    state AS "State",
    watt AS "Watt",
    group_dep AS "GroupDep",
    division AS "Division"
FROM assets;

CREATE OR REPLACE VIEW "ACCESSORIES" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    brand AS "Brand",
    description AS "Description",
    serial AS "Serial"
FROM accessories;

CREATE OR REPLACE VIEW "ASSETSLOGS" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    user_id AS "UserID",
    action AS "Action",
    action_logs AS "ActionLogs",
    date_modified AS "DateModified"
FROM assets_logs;

CREATE OR REPLACE VIEW "DEVICECONFIG" AS
SELECT
    device AS "Device",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9"
FROM device_config;

CREATE OR REPLACE VIEW "DISPOSAL" AS
SELECT
    unit_code AS "UnitCode",
    id_disposal AS "IDDisposal",
    start_date AS "StartDate",
    end_date AS "EndDate",
    status AS "Status",
    note AS "Note",
    user_id AS "UserID"
FROM disposal;

CREATE OR REPLACE VIEW "Documents" AS
SELECT
    document_id AS "DocumentID",
    reference_id AS "ReferenceID",
    user_stamp AS "UserStamp",
    date_time_stamp AS "DateTimeStamp",
    file_name AS "FileName",
    file_extension AS "FileExtension",
    file_description AS "FileDescription",
    file_binary AS "FileBinary"
FROM documents;

CREATE OR REPLACE VIEW "EmployeeTitle" AS
SELECT
    employee_code AS "EmployeeCode",
    title AS "Title"
FROM employee_title;

CREATE OR REPLACE VIEW "HANDOVER" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    handover_date AS "HandoverDate",
    user_created AS "UserCreated",
    form AS "Form",
    user_transfer AS "UserTransfer",
    department_transfer AS "DepartmentTransfer",
    office_transfer AS "OfficeTransfer",
    company_transfer AS "CompanyTransfer",
    area_transfer AS "AreaTransfer",
    type AS "Type",
    jira_code AS "JiraCode",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9"
FROM handover;

CREATE OR REPLACE VIEW "HANDOVEROTHERS" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    handover_date AS "HandoverDate",
    user_created AS "UserCreated",
    form AS "Form",
    user_a AS "UserA",
    department_a AS "DepartmentA",
    office_a AS "OfficeA",
    company_a AS "CompanyA",
    area_a AS "AreaA",
    user_b AS "UserB",
    department_b AS "DepartmentB",
    office_b AS "OfficeB",
    company_b AS "CompanyB",
    area_b AS "AreaB",
    type AS "Type",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9",
    state_a AS "StateA",
    state_b AS "StateB",
    status_a AS "StatusA",
    status_b AS "StatusB",
    handover_no AS "HandoverNo",
    jira_code AS "JiraCode"
FROM handover_others;

CREATE OR REPLACE VIEW "HISTORY" AS
SELECT
    id AS "ID",
    user_id AS "UserID",
    action AS "Action",
    date_time_ac AS "DateTimeAC"
FROM history;

CREATE OR REPLACE VIEW "IMAGEREPORT" AS
SELECT
    id AS "ID",
    type_report AS "TypeReport",
    image1 AS "Image1",
    image2 AS "Image2"
FROM image_report;

CREATE OR REPLACE VIEW "ITEMMASTERDATA" AS
SELECT
    id AS "ID",
    item AS "Item",
    tech_specs AS "TechSpecs"
FROM item_master_data;

CREATE OR REPLACE VIEW "KEYTECHSPECS" AS
SELECT
    name AS "Name",
    description_tech_specs AS "DescriptionTechSpecs"
FROM key_tech_specs;

CREATE OR REPLACE VIEW "PREFIXUNITCODE" AS
SELECT
    id_department AS "IDDepartment",
    prefix AS "Prefix",
    area AS "Area"
FROM prefix_unit_code;

CREATE OR REPLACE VIEW "PROBLEMREPORT" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    report_date AS "ReportDate",
    user_created AS "UserCreated",
    problem_summary AS "ProblemSummary",
    test_procedures AS "TestProcedures",
    solutions AS "Solutions",
    type AS "Type",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9",
    user_report AS "UserReport",
    department_report AS "DepartmentReport",
    office_report AS "OfficeReport",
    company_report AS "CompanyReport",
    area_report AS "AreaReport",
    problem_no AS "ProblemNo",
    jira_code AS "JiraCode"
FROM problem_report;

CREATE OR REPLACE VIEW "SERVICE" AS
SELECT
    main_assets AS "MainAssets",
    sub_assets AS "SubAssets",
    service_level AS "ServiceLevel",
    time_create_tag AS "TimeCreateTag"
FROM service;

CREATE OR REPLACE VIEW "SOFTWARE" AS
SELECT
    id AS "ID",
    unit_code AS "UnitCode",
    software AS "Software",
    version AS "Version"
FROM software;

CREATE OR REPLACE VIEW "UpdateEmployeeCode" AS
SELECT
    unit_code AS "UnitCode",
    hr_code AS "HRCode"
FROM update_employee_code;

CREATE OR REPLACE VIEW "COMBO" AS
SELECT
    combo_name AS "ComboName",
    type AS "Type",
    brand AS "Brand",
    detail1 AS "Detail1",
    detail2 AS "Detail2",
    detail3 AS "Detail3",
    detail4 AS "Detail4",
    detail5 AS "Detail5",
    detail6 AS "Detail6",
    detail7 AS "Detail7",
    detail8 AS "Detail8",
    detail9 AS "Detail9",
    model AS "Model",
    purchased_value AS "PurchasedValue",
    type_exp AS "TypeEXP",
    warranty_expiration AS "WarrantyExpiration"
FROM combo;

CREATE OR REPLACE VIEW "CheckingResult" AS
SELECT
    checking_result AS "CheckingResult"
FROM checking_result;

CREATE OR REPLACE VIEW "ASSETSAD" AS
SELECT
    id AS "id",
    operating_system AS "OperatingSystem",
    service_pack AS "ServicePack",
    version AS "Version",
    create_year AS "CreateYear",
    cn AS "CN",
    last_logon AS "LastLogon",
    status AS "Status"
FROM assets_ad;

CREATE OR REPLACE VIEW "ASSETSCHEKING" AS
SELECT
    checking_pc AS "CheckingPC",
    checking_result AS "CheckingResult",
    checking_action AS "CheckingAction",
    checking_remark AS "CheckingRemark"
FROM assets_checking;

CREATE OR REPLACE VIEW "ASSETSEMPATCHDETAIL" AS
SELECT
    id AS "id",
    sqnumber AS "sqnumber",
    os_language AS "os_language",
    install_error_code AS "install_error_code",
    bulletin_id AS "bulletin_id",
    update_name AS "update_name",
    patch_size AS "patch_size",
    resource_id_string AS "resource_id_string",
    uninstall_for AS "uninstall_for",
    patch_uninstall_status AS "patch_uninstall_status",
    service_pack AS "service_pack",
    patch_status AS "patch_status",
    patch_install_status_image AS "patch_install_status_image",
    patch_id AS "patch_id",
    patch_name AS "patch_name",
    patch_description AS "patch_description",
    vendor_name AS "vendor_name",
    superceded_by AS "superceded_by",
    branch_office_name AS "branch_office_name",
    platform_name AS "platform_name",
    severity AS "severity",
    patch_released_time AS "patch_released_time",
    resource_id AS "resource_id",
    os_name AS "os_name",
    customer_name AS "customer_name",
    resource_name AS "resource_name",
    user_id AS "user_id",
    managedcomputer_resource_id AS "managedcomputerResource_id",
    first_name AS "first_name",
    friendly_name AS "friendly_name",
    i18n AS "i18n",
    resourcemacip_resource_id AS "resourcemacipResource_id",
    patch_updated_time AS "patch_updated_time",
    patch_affected_status_label AS "patch_affected_status_label"
FROM assets_em_patch_detail;

CREATE OR REPLACE VIEW "ASSETSME" AS
SELECT
    id AS "Id",
    modernappdetails_remarks AS "modernappdetailsRemarks",
    processor_arch_name AS "processor_arch_name",
    previous_installation_status AS "previous_installation_status",
    remarks_args AS "remarks_args",
    agent_installed_on AS "agent_installed_on",
    agent_executed_on AS "agent_executed_on",
    os_platform_name AS "os_platform_name",
    resource_id_string AS "resource_id_string",
    service_pack AS "service_pack",
    processor_arch_type_id AS "processor_arch_type_id",
    processor_architecture AS "processor_architecture",
    os_platform AS "os_platform",
    agent_logged_on_users AS "agent_logged_on_users",
    osflavor_id AS "osflavor_id",
    search_tag AS "search_tag",
    status_label AS "status_label",
    error_kb_url AS "error_kb_url",
    full_name AS "full_name",
    owner_email_id AS "owner_email_id",
    operation_type AS "operation_type",
    agent_upgraded_on AS "agent_upgraded_on",
    installation_status AS "installation_status",
    error_code AS "error_code",
    computer_status_update_time AS "computer_status_update_time",
    managed_status AS "managed_status",
    branch_office_name AS "branch_office_name",
    agent_last_contact_time AS "agent_last_contact_time",
    description AS "description",
    domain_netbios_name AS "domain_netbios_name",
    fqdn_name AS "fqdn_name",
    mac_address AS "mac_address",
    owner AS "owner",
    os_version AS "os_version",
    last_sync_time AS "last_sync_time",
    ip_address AS "ip_address",
    agent_uninstalled_on AS "agent_uninstalled_on",
    agent_version AS "agent_version",
    agent_last_bootup_time AS "agent_last_bootup_time",
    resource_id AS "resource_id",
    os_name AS "os_name",
    location AS "location",
    computer_live_status AS "computer_live_status",
    customer_name AS "customer_name",
    resource_name AS "resource_name",
    customer_id AS "customer_id",
    remarks AS "remarks"
FROM assets_me;

CREATE OR REPLACE VIEW "ASSETSTM" AS
SELECT
    id AS "id",
    endpoint_id AS "endpointID",
    endpoint_host AS "endpointHost",
    endpoint_ip AS "endpointIP",
    endpoint_mac AS "endpointMAC",
    product AS "product",
    managing_server_id AS "managingServerID",
    ad_domain AS "adDomain",
    domain AS "domain",
    domain_hierarchy AS "domainHierarchy",
    logon_user AS "logonUser",
    platform AS "platform",
    client_program AS "clientProgram",
    connection_status AS "connectionStatus",
    isolation_status AS "isolationStatus",
    firewall AS "firewall",
    scan_method AS "scanMethod",
    update_agent AS "updateAgent",
    last_scheduled_scan_utc AS "lastScheduledScanUTC",
    last_manual_scan_utc AS "lastManualScanUTC",
    last_startup AS "lastStartup",
    last_connected AS "lastConnected",
    virus_scan_engine AS "virusScanEngine",
    virus_pattern AS "virusPattern",
    smart_scan_agent_pattern AS "smartScanAgentPattern",
    capabilities AS "capabilities",
    last_contact_time_category AS "LastContactTimeCategory"
FROM assets_tm;


-- Seed initial admin user with bcrypt hash of 'password123' (cost 10)
INSERT INTO user_info (user_id, pass, user_name, description, area, role, is_active)
VALUES ('admin', '$2a$10$2K8gASm.rDv.S383QJj4iOqy96DTQozGGkejEjtjJSACw98S.hGwO', 'System Administrator', 'Initial System Administrator', 'ALL', 'admin', TRUE)
ON CONFLICT (user_id) DO UPDATE SET pass = '$2a$10$2K8gASm.rDv.S383QJj4iOqy96DTQozGGkejEjtjJSACw98S.hGwO';

INSERT INTO permission (
    user_id, assets_modify, assets_del, department_modify, department_del,
    type_modify, type_del, user_manager, disposal, permission_approval,
    allow_see_all_from_allowed_areas
)
VALUES (
    'admin', TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE, TRUE,
    TRUE
)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO state_master_data (id, state, description) VALUES
    ('001', 'In Use', 'Thiết bị đang được sử dụng'),
    ('002', 'In Stock', 'Thiết bị lưu kho'),
    ('003', 'Broken', 'Thiết bị hỏng hóc'),
    ('004', 'Disposed', 'Thiết bị đã thanh lý'),
    ('005', 'Pending Transfer', 'Chờ chuyển giao')
ON CONFLICT (id) DO NOTHING;

INSERT INTO type (type, type_name, vn) VALUES
    ('DPC', 'Desktop PC', 'Máy tính để bàn'),
    ('LTP', 'Laptop', 'Máy tính xách tay'),
    ('MON', 'Monitor', 'Màn hình'),
    ('PRN', 'Printer', 'Máy in'),
    ('SVR', 'Server', 'Máy chủ'),
    ('NET', 'Network Device', 'Thiết bị mạng')
ON CONFLICT (type) DO NOTHING;

INSERT INTO key_tech_specs (name, description_tech_specs) VALUES
    ('CPU', 'Bộ vi xử lý'),
    ('RAM', 'Bộ nhớ trong'),
    ('SSD', 'Ổ cứng thể rắn'),
    ('HDD', 'Ổ cứng cơ'),
    ('VGA', 'Card đồ họa'),
    ('OS', 'Hệ điều hành')
ON CONFLICT (name) DO NOTHING;

INSERT INTO device_config (device, detail1, detail2, detail3, detail4, detail5, detail6, detail7, detail8, detail9) VALUES
    ('DPC', 'CPU', 'RAM', 'Storage 1', 'Storage 2', 'VGA', 'Mainboard', 'PSU', 'Case', 'OS'),
    ('LTP', 'CPU', 'RAM', 'Storage 1', 'Screen', 'VGA', 'Battery', 'Adapter', 'Weight', 'OS')
ON CONFLICT (device) DO NOTHING;
