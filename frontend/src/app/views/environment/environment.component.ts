import { Component, inject, signal, Signal, WritableSignal } from '@angular/core';
import { Environment, EnvironmentService } from '../../services/environment/environment.service';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { TabsModule } from 'primeng/tabs';
import { BadgeModule } from 'primeng/badge';

@Component({
  selector: 'app-environment',
  imports: [CommonModule, RouterModule, TableModule, ButtonModule, CardModule, TabsModule, BadgeModule],
  templateUrl: './environment.component.html',
  styleUrl: './environment.component.scss',
})
export class EnvironmentComponent {
  private envService = inject(EnvironmentService);
  private router = inject(Router)

  environments: WritableSignal<Environment[]>;
  loading = true;
  selectedEnvironment: WritableSignal<Environment>

  constructor() {
    this.environments = signal([])
    this.selectedEnvironment = signal({} as Environment)
  }

  ngOnInit() {
    this.envService.getEnvironments().subscribe({
      next: (data) => { this.environments.set(data); this.loading = false; },
      error: (err) => { console.error(err); this.loading = false; }
    });
  }

  goToEnvironment(env: Environment) {
    this.router.navigate(['/environments', env.id]);
  }

  createEnvironment() {
    // redirigir a formulario de creación o abrir modal
    this.router.navigate(['/environments/create']);
  }
}
