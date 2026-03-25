-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF to_regclass('pcnfsaid') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfsaid_dtsaida_codfilial ON pcnfsaid (dtsaida, codfilial)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfsaid_codcli ON pcnfsaid (codcli)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfsaid_codusur ON pcnfsaid (codusur)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfsaid_numtransvenda ON pcnfsaid (numtransvenda)';
  END IF;

  IF to_regclass('pcnfitem') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfitem_numtransvenda ON pcnfitem (numtransvenda)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcnfitem_codprod ON pcnfitem (codprod)';
  END IF;

  IF to_regclass('pcclient') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcclient_codcli ON pcclient (codcli)';
  END IF;

  IF to_regclass('pcfornec') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcfornec_codfornec ON pcfornec (codfornec)';
  END IF;

  IF to_regclass('pcprodut') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcprodut_codprod ON pcprodut (codprod)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcprodut_codfornec ON pcprodut (codfornec)';
  END IF;

  IF to_regclass('pcusuari') IS NOT NULL THEN
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_pcusuari_codusur ON pcusuari (codusur)';
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
  IF to_regclass('idx_pcnfsaid_dtsaida_codfilial') IS NOT NULL THEN
    EXECUTE 'DROP INDEX IF EXISTS idx_pcnfsaid_dtsaida_codfilial';
  END IF;
  EXECUTE 'DROP INDEX IF EXISTS idx_pcnfsaid_codcli';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcnfsaid_codusur';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcnfsaid_numtransvenda';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcnfitem_numtransvenda';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcnfitem_codprod';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcclient_codcli';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcfornec_codfornec';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcprodut_codprod';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcprodut_codfornec';
  EXECUTE 'DROP INDEX IF EXISTS idx_pcusuari_codusur';
END $$;
-- +goose StatementEnd
