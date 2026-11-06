import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { UserService } from '../../core/services/user.service';
import { environment } from '../../../environments/environment';
import { NzMessageService } from 'ng-zorro-antd/message';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NzFormModule,
    NzInputModule,
    NzButtonModule,
    NzIconModule
  ],
  templateUrl: './login.html',
  styleUrl: './login.less'
})
export class Login implements OnInit {
  loginForm!: FormGroup;
  loading = false;

  constructor(
    private fb: FormBuilder,
    private userService: UserService,
    private router: Router,
    private message: NzMessageService
  ) {}

  ngOnInit() {
    if (localStorage.getItem('currentUser')) {
      this.router.navigate(['/']);
      return;
    }

    this.loginForm = this.fb.group({
      username: ['', [Validators.required]],
      password: ['', [Validators.required]]
    });
  }

  onSubmit() {
    if (this.loginForm.valid) {
      this.loading = true;
      const { username, password } = this.loginForm.value;
      this.userService.login({ username, password }).subscribe({
        next: (res: any) => {
          this.loading = false;
          if (res.status && res.data) {
            localStorage.setItem('currentUser', JSON.stringify(res.data));
            this.message.success('Đăng nhập thành công!');
            this.router.navigate(['/']);
          } else {
            this.message.error(res.message || 'Tài khoản hoặc mật khẩu không chính xác.');
          }
        },
        error: (err) => {
          this.loading = false;
          this.message.error(err.error?.message || 'Không thể kết nối đến máy chủ.');
        }
      });
    } else {
      Object.values(this.loginForm.controls).forEach(control => {
        if (control.invalid) {
          control.markAsDirty();
          control.updateValueAndValidity({ onlySelf: true });
        }
      });
    }
  }

  loginWithSSO() {
    const ssoHost = environment.authConfig.host;
    const clientId = environment.authConfig.client_id;
    const callbackUrl = environment.authConfig.callback_path;

    if (ssoHost.includes('9091') || ssoHost.includes('authelia') || ssoHost.includes('127.0.0.1')) {
      const state = Array.from(window.crypto.getRandomValues(new Uint8Array(16)))
        .map(b => b.toString(16).padStart(2, '0'))
        .join('');
      window.location.href = `${ssoHost}/api/oidc/authorization?client_id=${clientId}&redirect_uri=${encodeURIComponent(callbackUrl)}&response_type=code&scope=openid%20profile%20email&state=${state}`;
    } else {
      window.location.href = `${ssoHost}?clientId=${clientId}&callbackUrl=${callbackUrl}`;
    }
  }
}
