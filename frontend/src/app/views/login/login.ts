import { Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Auth } from '../../services/auth/auth';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, CommonModule],
  templateUrl: './login.html',
  styleUrl: './login.css',
})
export class Login {
  private router = inject(Router)
  private fb = inject(FormBuilder)
  private authService = inject(Auth)

  loginForm: FormGroup

  constructor() {
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required]
    })
  }

  onSubmit() {
    console.log('this.loginForm.valid', this.loginForm.valid)
    if (this.loginForm.valid) {
      this.authService.login({email: this.loginForm.value.email, password: this.loginForm.value.password}).subscribe({
        next: () => this.router.navigate(['/dashboard/environments']),
      })
    }
  }

  get f() {
    return this.loginForm.controls;
  }
}
