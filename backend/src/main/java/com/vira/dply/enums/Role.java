package com.vira.dply.enums;

public enum Role {

  /** 
   * Control total:
   * - gestionar usuarios
   * - gestionar teams
   * - gestionar entornos (clusters)
   * - todos los permisos de deploy
   */
  OWNER,

  /**
   * Admin del team:
   * - crear/actualizar deployments
   * - escalar, configurar recursos
   * - NO puede borrar el entorno ni gestionar usuarios globales
   */
  ADMIN,

  /**
   * Developer:
   * - deploy
   * - update
   * - scale
   * - ver logs
   */
  DEVELOPER,

  /**
   * Solo lectura:
   * - ver estado
   * - ver logs
   */
  VIEWER,

  MEMBER
}