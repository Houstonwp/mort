export interface ClassifiedValue {
  code: string;
  label: string;
}

export interface AxisDefinition {
  id: string;
  scaleType: ClassifiedValue;
  axisName: string;
  minValue: string;
  maxValue: string;
  increment: string;
}

export interface TableMeta {
  scalingFactor?: string;
  dataType?: ClassifiedValue;
  nation?: ClassifiedValue;
  tableDescription?: string;
  axes?: AxisDefinition[];
}

export interface RateEntry {
  age: number;
  duration?: number | null;
  rate?: number | null;
}

export interface TablePayload {
  index: number;
  metadata?: TableMeta;
  rates?: RateEntry[];
}

export interface ClassificationPayload {
  tableIdentity: string;
  providerDomain?: string;
  providerName?: string;
  tableReference?: string;
  contentType?: ClassifiedValue;
  tableName?: string;
  tableDescription?: string;
  comments?: string;
  keywords?: string[];
}

export interface ConvertedTable {
  identifier: string;
  version?: string;
  classification?: ClassificationPayload;
  tables?: TablePayload[];
}

export interface TableSummary {
  identifier: string;
  tableIdentity: string;
  name: string;
  provider: string;
  summary: string;
  keywords: string[];
  version: string;
  detailPath: string;
}

export interface TableIndexEntry extends TableSummary {
  filePath: string;
}

export type DetailTab = 'classification' | 'metadata' | 'rates';
