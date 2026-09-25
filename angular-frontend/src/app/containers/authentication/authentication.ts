import { Component, OnInit, NgZone } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { UserService } from '../../core/services/user.service';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzSpinModule } from 'ng-zorro-antd/spin';

@Component({
  selector: 'app-authentication',
  standalone: true,
  imports: [CommonModule, NzSpinModule],
  template: `
    <div style="height: 100vh; display: flex; flex-direction: column; justify-content: center; align-items: center; background: #f0f2f5;">
      <nz-spin nzSimple nzSize="large"></nz-spin>
      <h3 style="margin-top: 16px; font-weight: 500; color: #1890ff;">Xác thực tài khoản với SSO Portal...</h3>
    </div>
  `
})
export class Authentication implements OnInit {
  constructor(
    private activatedRoute: ActivatedRoute,
    private userService: UserService,
    private router: Router,
    private message: NzMessageService,
    private ngZone: NgZone
  ) {}

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      if (params['error']) {
        const errorMsg = params['error_description'] || params['error'];
        this.message.error(`Lỗi SSO: ${errorMsg}`, { nzDuration: 8000 });
        return;
      }

      const authCode = params['code'] || params['authCode'] || '';
      if (!authCode) {
        this.message.error('Không tìm thấy mã xác thực (AuthCode) từ SSO Portal.');
        setTimeout(() => this.router.navigate(['/login']), 2000);
        return;
      }
      this.verify(authCode);
    });
  }

  verify(authCode: string) {
    this.userService.verifyToken(authCode).subscribe({
      next: (res: any) => {
        if (res.status && res.data) {
          localStorage.setItem('currentUser', JSON.stringify(res.data));
          this.message.success('Đăng nhập thành công!');
          this.ngZone.run(() => this.router.navigateByUrl('/')).then();
        } else {
          const msg = res.message || 'Xác thực tài khoản thất bại hoặc tài khoản bị khóa.';
          this.message.error(msg, { nzDuration: 5000 });
          setTimeout(() => {
            this.router.navigate(['/login']);
          }, 5000);
        }
      },
      error: (err) => {
        console.error(err);
        const errorMsg = err.error?.message || 'Lỗi hệ thống khi xác thực tài khoản hoặc tài khoản Inactive.';
        this.message.error(errorMsg, { nzDuration: 5000 });
        setTimeout(() => {
          this.router.navigate(['/login']);
        }, 5000);
      }
    });
  }
}
