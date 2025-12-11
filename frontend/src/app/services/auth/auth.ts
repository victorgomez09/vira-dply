import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';
import { Observable, tap } from 'rxjs';

type Login = {
  email: string;
  password: string;
}

type AuthResponse = {
  token: string;
  username: string;
  roles: string[];
}

@Injectable({
  providedIn: 'root',
})
export class Auth {
  private http = inject(HttpClient)
  private baseUrl = `${environment.api.baseUrl}/auth`

  login(data: Login): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.baseUrl}/login`, data).pipe(tap(res => localStorage.setItem('token', res.token)))
  }

  register(data: { username: string; password: string }): Observable<any> {
    return this.http.post(`${this.baseUrl}/register`, data);
  }

  logout(): void {
    localStorage.removeItem('token');
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }
}
