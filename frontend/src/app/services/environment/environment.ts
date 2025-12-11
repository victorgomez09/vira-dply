import { inject, Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export type EnvironmentData = {
  id: string,
  name: string,
  kubeContext: string,
  kubeConfigPath: string,
  teams: any[]
}

@Injectable({
  providedIn: 'root',
})
export class Environment {
  private http = inject(HttpClient)
  private baseUrl = `${environment.api.baseUrl}/api/environments`;

  getAll(): Observable<EnvironmentData[]> {
    return this.http.get<EnvironmentData[]>(this.baseUrl);
  }
}
