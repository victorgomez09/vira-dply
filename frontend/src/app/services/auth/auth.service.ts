import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { tap } from 'rxjs/operators';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface AuthResponse {
  token: string;
  username: string;
  roles: string[];
}

@Injectable({
  providedIn: 'root' // sigue siendo singleton
})
export class AuthService {
  private api = `${environment.API_URL}/api/auth`;

  constructor(private http: HttpClient) { }

  login(credentials: { username: string; password: string }): Observable<AuthResponse> {
    return this.http.post<AuthResponse>(`${this.api}/login`, credentials)
      .pipe(
        tap(res => localStorage.setItem('token', res.token))
      );
  }

  register(data: { username: string; password: string }): Observable<any> {
    return this.http.post(`${this.api}/register`, data);
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
