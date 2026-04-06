import { IsString, IsNumber, IsObject, IsOptional, MinLength, MaxLength } from 'class-validator';

export class IngestMetricDto {
  @IsString()
  @MinLength(1)
  @MaxLength(128)
  name: string;

  @IsNumber()
  value: number;

  @IsObject()
  @IsOptional()
  tags?: Record<string, string>;
}
