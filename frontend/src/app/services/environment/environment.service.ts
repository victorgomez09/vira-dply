import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface Environment {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
}

@Injectable({
  providedIn: 'root'
})
export class EnvironmentService {
  private api = `${environment.API_URL}/api/environments`;
  private http = inject(HttpClient)

  getEnvironments(): Observable<Environment[]> {
    return this.http.get<Environment[]>(this.api);
  }
}
