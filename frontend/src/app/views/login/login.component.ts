import { Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { ToastModule } from 'primeng/toast';
import { MessageModule } from 'primeng/message';

import { AuthService } from '../../services/auth/auth.service';
import { CardModule } from 'primeng/card';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, InputTextModule, ButtonModule, ToastModule, MessageModule, CardModule],
  providers: [MessageService],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss',
})
export class LoginComponent {
  private fb = inject(FormBuilder)
  private authService = inject(AuthService)
  private router = inject(Router)
  private messageService = inject(MessageService)

  loginForm!: FormGroup;
  formSubmitted = false;

  ngOnInit(): void {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required],
    });
  }

  onSubmit() {
    if (this.loginForm.invalid) return;

    this.authService.login(this.loginForm.value).subscribe({
      next: () => this.router.navigate(['/environments']),
      error: err => this.messageService.add({ severity: 'error', summary: 'Login Failed', detail: err.error?.message })
    });
  }

  isInvalid(controlName: string) {
    const control = this.loginForm.get(controlName);
    return control?.invalid && (control.touched || this.formSubmitted);
  }
}
