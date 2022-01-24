import pytest

from .. models import *


def test_model(dbsession):

    # fx
    eur = Forex(code='EUR', units_per_usd=0.921, usd_per_unit=1.085)
    gbp = Forex(code='GBP', units_per_usd=0.773, usd_per_unit=1.292)

    
    # vendors
    acme_corp = Vendor(name='acme-corp', currency='USD', 
                             discount_code='abc123', discount=0.1)
    amazon = Vendor(name='amazon', currency='GBP')
    
    # brands
    yamaha = Brand(name='yamaha')
    samsung = Brand(name='samsung')
    microsoft = Brand(name='microsoft')

    # products
    product_a = Product(name='productA', brand=yamaha)
    product_b = Product(name='productB', brand=samsung)
    product_c = Product(name='productC', brand=microsoft)

    # price quotes
    quote1 = Price(product=product_a, vendor=acme_corp, currency='USD', value=280)
    quote2 = Price(product=product_a, vendor=amazon, currency='GBP', value=210)

    # relationships
    acme_corp.brands = [yamaha, samsung]
    amazon.brands = [yamaha, samsung, microsoft]

    # session
    session = dbsession

    acme_corp.add_product(session, 'sony', 'walkman', 99.99)
    pc = mk_adder(session, acme_corp)
    ss = mk_adder(session, amazon)
    session.add_all([
        eur, gbp, 
        acme_corp, amazon, 
        yamaha, samsung, microsoft, 
        product_a, product_b, product_c,
        quote1, quote2
    ])
    session.flush()
    # session.commit()

    assert product_a.name == 'productA'

    assert 'walkman' in [p.name for p in session.query(Product).all()]
